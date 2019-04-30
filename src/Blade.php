<?php
/**
 * Created by PhpStorm.
 * User: chotow
 * Date: 2018/9/27
 * Time: 8:55 PM
 */

namespace Choyri\Gamyeewai;

use Closure;
use Exception;
use Choyri\Gamyeewai\Exceptions\DingTalkException;
use Choyri\Gamyeewai\Exceptions\GamyeewaiException;
use Illuminate\Contracts\Debug\ExceptionHandler;
use Nord\Lumen\ChainedExceptionHandler\ChainedExceptionHandler;

class Blade
{
    public static $env;

    protected $config;
    protected $reporter;

    public function __construct(array $config)
    {
        $this->config = $config;
        $this->reporter = new DingTalkReporter($config);
    }

    /**
     * 处理异常
     *
     * @param Exception $e
     * @param object|Closure $userInfo
     * @param bool $sendDingTalk
     *
     * @throws GamyeewaiException
     */
    public function report(Exception $e, $userInfo = null, bool $sendDingTalk = true): void
    {
        $handle = app(ExceptionHandler::class);

        // 兼容 nordsoftware/lumen-newrelic
        if ($handle instanceof ChainedExceptionHandler) {
            $handle = new \App\Exceptions\Handler();
        }

        if (!$handle->shouldReport($e) || $e instanceof DingTalkException || $e instanceof GamyeewaiException) {
            return;
        }

        try {
            $logger = app('Psr\Log\LoggerInterface');
        } catch (Exception $e) {
            throw new GamyeewaiException('Logger 实例不存在', 0, $e);
        }

        $logger->error($e->getMessage(), LogstashFormatter::generateInfo($e, $userInfo));

        if ($sendDingTalk) {
            $this->reporter->handle($e);
        }
    }

    /**
     * 发动钉钉机器人消息
     * 内容格式见 http://t.cn/RYV10UW
     *
     * @param array $data
     * @param string|null $token
     */
    public function sendDingTalk(array $data, string $token = null): void
    {
        $this->reporter->send($data, $token);
    }

    /**
     * 发动钉钉机器人文本消息
     *
     * @param string $content
     * @param string|null $token
     */
    public function sendTextDingTalk(string $content, string $token = null): void
    {
        $this->reporter->sendText($content, $token);
    }

    /**
     * 伪装环境
     *
     * @param string $env
     *
     * @return $this
     */
    public function fakeEnv(string $env): Blade
    {
        self::$env = $env;

        return $this;
    }
}
