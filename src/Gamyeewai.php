<?php
/**
 * Created by PhpStorm.
 * User: chotow
 * Date: 2018/9/27
 * Time: 8:59 PM
 */

namespace Choyri\Gamyeewai;

use Closure;
use Exception;
use Illuminate\Support\Facades\Facade as LaravelFacade;

/**
 * @method static void report(Exception $e, object|Closure $userInfo = null, bool $sendDingTalk = true)
 * @method static void sendDingTalk(array $data, string $token = null)
 * @method static void sendTextDingTalk(string $content, string $token = null)
 * @method static Blade fakeEnv(string $env)
 *
 * @see \Choyri\Gamyeewai\Blade
 */
class Gamyeewai extends LaravelFacade
{
    public static function getFacadeAccessor()
    {
        return 'gamyeewai';
    }
}
