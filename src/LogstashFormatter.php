<?php
/**
 * Created by PhpStorm.
 * User: chotow
 * Date: 2018/9/21
 * Time: 上午10:18
 */

namespace Choyri\Gamyeewai;

use Closure;
use Exception;
use Firebase\JWT\JWT;
use Illuminate\Support\Facades\Auth;

class LogstashFormatter extends \Monolog\Formatter\LogstashFormatter
{
    use Helpers;

    public static $baseMeta;
    public static $token;
    public static $userInfo;

    public function format(array $record)
    {
        $record['channel'] = 'gamyeewai';

        // 预置字段
        $record['context'] = array_merge($record['context'] ?? [], [
            'env' => Helpers::getRealEnv(),
        ]);

        return parent::format($record);
    }

    /**
     * 生成带有详细信息的日志
     *
     * @param Exception $e
     * @param object|Closure $userInfo
     * @return array
     */
    public static function generateInfo(Exception $e, $userInfo = null): array
    {
        if (!Helpers::isProdEnv()) {
            return Helpers::isLocalEnv() ? ['exception' => $e] : [];
        }

        $request = app('request');
        $content = $request->getContent();

        $startTime = $_SERVER['REQUEST_TIME_FLOAT'] ?? microtime(true);

        self::$token = $request->input('token') ?: $request->bearerToken() ?: $request->cookie('token') ?: null;

        try {
            $secret = config('jwt.secret', env('JWT_SECRET'));
            $algorithm = config('jwt.algorithm', env('JWT_ALGORITHM', 'HS256'));
            $payload = JWT::decode(self::$token, $secret, [$algorithm]);
        } catch (Exception $_) {
            $payload = '解码失败';
        }

        return array_merge(self::getBaseData(), [
            'client_ip'        => $_SERVER['HTTP_REMOTEIP'] ?? $request->getClientIp(),
            'headers'          => $request->headers->all(),
            'body'             => empty($content) ? 'n/a' : $content,
            'exec_time'        => intval((microtime(true) - $startTime) * 1000) . 'ms',
            'token'            => self::$token,
            'token_payload'    => $payload,
            'user'             => self::getUserInfo($userInfo),
            'message_location' => $e->getFile() . ':' . $e->getLine(),
            'message_trace'    => $e->getTraceAsString(),
        ]);
    }

    /**
     * 获取基本元信息
     *
     * @return array
     */
    public static function getBaseData(): array
    {
        if (self::$baseMeta) {
            return self::$baseMeta;
        }

        $request = app('request');

        return self::$baseMeta = [
            'timestamp' => time(),
            'url'       => $request->url(),
            'method'    => $request->method(),
            'input'     => $request->input() ?: 'n/a',
        ];
    }

    /**
     * 获取用户信息
     *
     * @param object|Closure $userInfo
     * @return array
     */
    public static function getUserInfo($userInfo = null): array
    {
        if (isset(self::$userInfo)) {
            return self::$userInfo;
        }

        if ($userInfo instanceof Closure) {
            return self::$userInfo = $userInfo();
        }

        try {
            if (is_object($userInfo)) {
                $user = $userInfo;
            } else {
                if (!self::$token || !class_exists(Auth::class) || !($user = Auth::user())) {
                    return ['n/a'];
                }
            }

            return self::$userInfo = [
                'id'      => $user->id ?? 'n/a',
                'name'    => $user->username ?? $user->nickname ?? 'n/a',
                'unionid' => $user->unionid ?? 'n/a',
                'openid'  => $user->openid ?? 'n/a',
            ];
        } catch (Exception $_) {
            return ['用户信息获取失败 ' . $_->getMessage()];
        }
    }
}
