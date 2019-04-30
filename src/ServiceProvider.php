<?php
/**
 * Created by PhpStorm.
 * User: chotow
 * Date: 2018/9/27
 * Time: 8:41 PM
 */

namespace Choyri\Gamyeewai;

use Illuminate\Database\Events\QueryExecuted;
use Illuminate\Foundation\Application as LaravelApplication;
use Illuminate\Support\ServiceProvider as ParentServiceProvider;
use Laravel\Lumen\Application as LumenApplication;

class ServiceProvider extends ParentServiceProvider
{
    use Helpers;

    public function boot()
    {
        if (Helpers::isLocalEnv()) {
            $this->recordSQLs();
        }
    }

    public function register()
    {
        $this->disposeLoggingConfig();

        $this->app->singleton('gamyeewai', function ($app) {
            return new Blade($app['config']['gamyeewai']);
        });
    }

    private function recordSQLs(): void
    {
        $this->app->make('db')->listen(function (QueryExecuted $query) {
            $sqlWithPlaceholders = str_replace(['%', '?'], ['%%', '%s'], $query->sql);

            $bindings = $query->connection->prepareBindings($query->bindings);
            $bindings = array_map([$query->connection->getPdo(), 'quote'], $bindings);

            $duration = $query->time . 'ms';
            $realSQL = vsprintf($sqlWithPlaceholders, $bindings);

            $content = sprintf('[%s] %s', $duration, $realSQL);

            $this->app->make('log')->channel('sql')->info($content);
        });
    }

    private function disposeLoggingConfig(): void
    {
        $source = realpath(__DIR__ . '/config.php');

        if ($this->app instanceof LaravelApplication && $this->app->runningInConsole()) {
            $this->publishes([$source => config_path('logging.php')]);
        } elseif ($this->app instanceof LumenApplication) {
            $this->app->configure('logging');
        }

        $source = require $source;
        $current = $this->app['config']->get('logging', []);
        $custom = $this->app->basePath('config') . '/logging.php';

        $finial = file_exists($custom) ? array_merge($source, $current) : array_merge($current, $source);

        $autoLocal = $this->app['config']['gamyeewai.auto_local_for_sql'] ?? true;

        if ($autoLocal && Helpers::isLocalEnv()) {
            $finial['default'] = 'local';
        }

        $this->app['config']->set('logging', $finial);
    }
}
