# 锦衣卫

当初写的一个内部用的 composer 包，现不再维护。

---

- 对错误异常进行记录。
  - 线上会以 JSON 格式记录日志，可在 Kibana 中添加后查看。
  - 如果配置了钉钉机器人 token，会自动发送报警消息。
- 本地开发时，默认会记录所有执行的 SQL。
  - `storage/logs/gamyeewai-sql.log`
- 可使用 `Gamyeewai::sendDingTalk` 主动发送钉钉机器人消息。
- 可使用 `Gamyeewai::sendTextDingTalk` 主动发送钉钉机器人文本消息。
- 可使用 `Gamyeewai::fakeEnv` 模拟应用环境。

默认状态下会自动检测环境。处于 `local` 开发环境下日志会输出到 `gamyeewai-local.log` （正常格式），相当于设置环境变量 `LOG_CHANNEL=local`；其他环境下会输出到 `gamyeewai-YYYY-MM-DD.log` （JSON 格式）。可通过配置项关闭该特性。


## 框架要求

Laravel/Lumen >= 5.5


## 安装

```shell
composer require choyri/gamyeewai
```


## 配置

#### 第一步，Laravel 和 Lumen 有区别。

- Laravel

    替换默认日志配置文件 `logging.php`：

    ```shell
    php artisan vendor:publish --provider="Choyri\Gamyeewai\ServiceProvider"
    ``` 

- Lumen

    注册 ServiceProvider：

    ```php
    // bootstrap/app.php

    $app->register(Choyri\Gamyeewai\ServiceProvider::class);
    ```

    默认会使用扩展包下的 `config.php` 配置文件。如果需要自定义，将 `vendor/choyri/gamyeewai/src/config.php` 复制到 `项目根目录/config` 目录下，并将文件名改为 `logging.php`。


#### 第二步，接管 report。

```php
// app/Exceptions/Handler.php

...

public function report(Exception $e)
{
    // parent::report($e);
    Gamyeewai::report($e); // 不要忘记命名空间引用
}

...
```

如果遇到 500 错误，但内容为空——该扩展包自己报错了，可以考虑这样子：

```php
try {
    Gamyeewai::report($e);
} catch (Exception $_e) {
    parent::report($_e);
}
```

`report` 可以传入第二个参数，用于指定当前用户信息（来源）：

```php
// 可以传入一个对象
$user = new stdClass();
$user->id = 1;
$user->nickname = 'zhangsan';
$user->unionid = 'xxx';
$user->openid = 'yyy';
Gamyeewai::report($e, $user);

// 可以传入一个闭包
Gamyeewai::report($e, function () {
    return [
        'id'       => 2,
        'nickname' => 'lisi',
    ];
});
```

`report` 还传入第三个参数，用于控制是否发送钉钉报警消息。


#### 第三步，配置项。

不存在 `dingtalk_token` 时，机器人不会发报警消息。不存在 `kibana_xxx` 时，报警信息不会携带「查看详情」的 Kibana 链接。

```php
// app/config/gamyeewai.php

return [
    // 钉钉机器人 token
    'dingtalk_token' => 'xxx',

    // 钉钉机器人重复报警阈值 一分钟内次数超过此值机器人会艾特所有人 忽视此项时默认值为10
    // 'dingtalk_threshold' => 10,

    // kibana 的域名
    'kibana_domain' => 'https://elk.exmaple.com',

    // kibana 的 index  可从对应项目的分享链接中找到
    'kibana_index' => 'xxx',

    // 自动辨别本地模式 # 默认为 ture  启用后本地开发时即使 LOG_CHANNEL 不为 local 也会自动设为 local 
    // 'auto_local_for_sql' => true,
];
```

这里有一个鸡生蛋还是蛋生鸡的问题。`kibana_index` 需要在添加索引后才能通过分享链接看到，而添加索引需要先生成日志文件。如果使用该扩展包生成日志，它又需要配置 `kibana_index`。

**所以，这一步的配置不是必须的；只接管 report 不会报错。**这样子，就能生成日志文件了；然后你可以在下一版再进行这一步，或者，手动把 JSON 格式的日志放到线上，然后去 Kibana 里添加索引。


## 其他函数

#### Gamyeewai::sendDingTalk(array $data, string $token = null)

主动发送钉钉机器人消息。消息格式见 [钉钉文档](http://t.cn/RYV10UW)。

该函数可以手动指定 `token`。

如果不传入指定的 `token`，又不配置 `dingtalk_token`，调用该方法是会报错的，😅。

```php
Gamyeewai::sendDingTalk([
    'msgtype'  => 'markdown',
    'markdown' => [
        'title' => '我是标题',
        'text'  => '我是内容',
    ],
]);
```

#### Gamyeewai::sendTextDingTalk(string $content, string $token = null)

主动发送钉钉机器人文本消息。

该函数可以手动指定 `token`。

如果不传入指定的 `token`，又不配置 `dingtalk_token`，调用该方法是会报错的，😅。

```php
Gamyeewai::sendTextDingTalk('我是内容');
```

#### Gamyeewai::fakeEnv(string $env)

伪装当前应用环境。比如，本地开发环境下可以模拟生产环境以生成 JSON 格式的日志，或者产生报警。

```php
// app/Exceptions/Handler.php

...
Gamyeewai::fakeEnv('prod')->report($e);
...
```

---

> 🌿 煎茶坐看梨门雨
