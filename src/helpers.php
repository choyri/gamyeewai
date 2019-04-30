<?php
/**
 * Created by PhpStorm.
 * User: chotow
 * Date: 2018/10/8
 * Time: 2:29 PM
 */

namespace Choyri\Gamyeewai;

use Illuminate\Support\Str;

trait Helpers
{
    private static $env;

    public static function getRealEnv(): string
    {
        return env('APP_REAL_ENV', config('app.env', env('APP_ENV', 'n/a')));
    }

    public static function isProdEnv(): bool
    {
        return strpos('production', self::getEnv()) !== false;
    }

    public static function isLocalEnv(): bool
    {
        return self::getEnv() === 'local';
    }

    public static function getAppName(): string
    {
        $groupName = Str::slug(env('CI_PROJECT_NAMESPACE', ''), '_');
        $projectName = Str::slug(env('CI_APP_NAME', ''), '_');
        $appName = $groupName . '-' . $projectName;

        return $appName !== '-' ? $appName : env('APP_NAME', 'defaultAppName');
    }

    private static function getEnv(): string
    {
        self::$env || self::$env = self::getRealEnv();

        return Blade::$env ?? self::$env;
    }
}
