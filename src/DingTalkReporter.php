<?php
/**
 * Created by PhpStorm.
 * User: chotow
 * Date: 2018/9/20
 * Time: 下午2:13
 */

namespace Choyri\Gamyeewai;

use Exception;
use Choyri\Gamyeewai\Exceptions\DingTalkException;
use GuzzleHttp\Client;

class DingTalkReporter
{
    use Helpers;

    const COUNT_CACHE_KEY = 'dingtalkReporter_count';
    const HOST = 'https://oapi.dingtalk.com/robot/send?access_token=';

    private $client;
    protected $config;
    protected $title;

    public function __construct(array $config)
    {
        $this->config = $config;
    }

    /**
     * 处理异常 然后报警 🌝
     *
     * @param Exception $e
     */
    public function handle(Exception $e): void
    {
        if (!Helpers::isProdEnv() || !isset($this->config['dingtalk_token'])) {
            print_r('??');
            return;
        }

        $this->title = $e->getMessage() ?: '不知道是啥错误';

        $data = [
            'msgtype'  => 'markdown',
            'markdown' => [
                'title' => $this->title,
                'text'  => $this->generateMessage($e),
            ],
        ];

        $this->send($data);
        $this->checkRepeatError();
    }

    /**
     * 发送钉钉机器人消息
     *
     * @param array $data
     * @param string|null $token
     */
    public function send(array $data, string $token = null): void
    {
        $resp = null;

        try {
            $this->getClient($token)->post('', ['json' => $data])->getBody()->getContents();
        } catch (Exception $e) {
            app('log')->error($e->getMessage());
        }
    }

    /**
     * 发送钉钉机器人文本消息
     *
     * @param string $content
     * @param string|null $token
     *
     */
    public function sendText(string $content, string $token = null): void
    {
        $this->send([
            'msgtype' => 'text',
            'text'    => [
                'content' => $content,
            ],
        ], $token);
    }

    /**
     * 检测重复错误
     */
    private function checkRepeatError(): void
    {
        $cacheKey = join('#', [Helpers::getAppName(), self::COUNT_CACHE_KEY]);

        $now = date('YmdHi', time());

        $previousData = app('cache')->get($cacheKey, [$now => 0]);

        $count = $previousData[$now] ?? 0;
        ++$count;

        $previousData[$now] = $count;

        app('cache')->put($cacheKey, $previousData, 1);

        if ($count < ($this->config['dingtalk_threshold'] ?? 10)) {
            return;
        }

        $data = [
            'msgtype' => 'text',
            'text'    => [
                'content' => "[捂脸哭][捂脸哭][捂脸哭] 1 分钟内连报 $count 次错，快来人呐！",
            ],
            'at'      => [
                'isAtAll' => true,
            ],
        ];

        $this->send($data);
    }

    /**
     * 处理异常信息 返回报警内容
     *
     * @param Exception $e
     *
     * @return string
     */
    private function generateMessage(Exception $e): string
    {
        $messages = array_merge([
            'message' => $this->title,
        ], LogstashFormatter::getBaseData(), [
            'user'     => LogstashFormatter::getUserInfo(),
            'location' => $e->getFile() . ':' . $e->getLine(),
            'trace'    => $this->disposeTraceString($e->getTraceAsString()),
        ]);

        $timestamp = tap($messages['timestamp'], function ($value) use (&$messages) {
            // 让时间戳变可读
            $messages['timestamp'] = date('Y-m-d H:i:s', $value);
        });

        // 把 method 和 url 放一行  然后删掉 method
        $messages['url'] = join(' ', [$messages['method'], $messages['url']]);
        unset($messages['method']);

        $ret = '';

        foreach ($messages as $key => $item) {
            // 不处理堆栈信息
            if ($key !== 'trace') {
                // JSON 编码后 删除头尾的双引号
                $item = trim(json_encode($item, JSON_UNESCAPED_SLASHES + JSON_UNESCAPED_UNICODE), '"');
            }

            $ret .= join("\n", ["### $key", '```', $item, '```', '---', '']);
        }

        $ret .= $this->disposeKibanaLink($timestamp);

        return $ret;
    }

    /**
     * 处理异常堆栈信息
     *
     * @param string $trace
     *
     * @return string
     */
    private function disposeTraceString(string $trace): string
    {
        // 从堆栈信息里取前 3 行
        $arrayString = array_slice(explode("\n", $trace), 0, 3);

        // 去掉路径里多余的前半部分 # /var/www/xxx/backend/vendor => /vendor
        $arrayString = preg_replace('/ .*?\/vendor/', ' /vendor', $arrayString);

        array_push($arrayString, '...');

        return join("\n", $arrayString);
    }

    /**
     * 获取此次异常对应的 Kibana 链接地址
     *
     * @param int $timestamp
     *
     * @return null|string
     */
    private function disposeKibanaLink(int $timestamp): string
    {
        $domain = $this->config['kibana_domain'] ?? null;
        $index = $this->config['kibana_index'] ?? null;

        if (!$domain || !$index) {
            return '';
        }

        $domain = rtrim(trim($domain), '/');

        $url = "{$domain}/app/kibana#/discover?_g=(time:(from:now-3d,mode:quick,to:now))&_a=(filters:!((query:(match:(timestamp:(query:'{$timestamp}'))))),index:'{$index}')";

        // 链接里有括号 得转义一下 否则钉钉的阉割版 Markdown 不识别
        $url = str_replace(['(', ')'], ['%28', '%29'], $url);

        return "[查看详情]($url)";
    }

    /**
     * 获取 Guzzle 客户端
     *
     * @param string|null $token
     *
     * @return Client
     * @throws DingTalkException
     */
    private function getClient(string $token = null): Client
    {
        $_getClient = function (string $token) {
            return new Client([
                'base_uri' => self::HOST . $token,
                'timeout'  => 3,
            ]);
        };

        if ($token) {
            return $_getClient($token);
        }

        $token = $this->config['dingtalk_token'] ?? null;

        if (!$token) {
            throw new DingTalkException('Access Token 不存在');
        }

        $this->client || $this->client = $_getClient($token);

        return $this->client;
    }
}
