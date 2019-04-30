<?php
/**
 * Created by PhpStorm.
 * User: chotow
 * Date: 2018/9/21
 * Time: 上午10:17
 */

namespace Choyri\Gamyeewai;

use Illuminate\Log\Logger;

class LogstashHandler
{
    public function __invoke(Logger $logger)
    {
        foreach ($logger->getHandlers() as $handler) {
            $handler->setFormatter(new LogstashFormatter(Helpers::getAppName(), null, '', '', \Monolog\Formatter\LogstashFormatter::V1));
        }
    }
}
