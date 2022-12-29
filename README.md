<p align="center">
	<img alt="logo" src="https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/logo.svg">
</p>
<h1 align="center" style="margin: 30px 0 30px; font-weight: bold;">OpenSCA-Cli</h1>
<p align="center">
	<a href="https://github.com/XmirrorSecurity/OpenSCA-cli/blob/master/LICENSE"><img src="https://img.shields.io/github/license/XmirrorSecurity/OpenSCA-cli?style=flat-square"></a>
	<a href="https://github.com/XmirrorSecurity/OpenSCA-cli/releases"><img src="https://img.shields.io/github/v/release/XmirrorSecurity/OpenSCA-cli?style=flat-square"></a>
</p>

## 项目介绍

**OpenSCA** 用来扫描项目的第三方组件依赖及漏洞信息。

---

## 检测能力

`OpenSCA`现已支持以下编程语言相关的配置文件解析及对应的包管理器，后续会逐步支持更多的编程语言，丰富相关配置文件的解析。

| 支持语言     | 包管理器   | 解析文件                                                     |
| ------------ | ---------- | ------------------------------------------------------------ |
| `Java`       | `Maven`    | `pom.xml`                                                    |
| `Java`       | `Gradle`   | `.gradle` `.gradle.kts`                                      |
| `JavaScript` | `Npm`      | `package-lock.json` `package.json` `yarn.lock`               |
| `PHP`        | `Composer` | `composer.json` `composer.lock`                              |
| `Ruby`       | `gem`      | `gemfile.lock`                                               |
| `Golang`     | `gomod`    | `go.mod` `go.sum`                                            |
| `Rust`       | `cargo`    | `Cargo.lock`                                                 |
| `Erlang`     | `Rebar`    | `rebar.lock`                                                 |
| `Python`     | `Pip`      | `Pipfile` `Pipfile.lock` `setup.py` `requirements.txt` `requirements.in` （有的解析会对python版本有要求，后两者的解析需要具备pipenv环境，需要联网。） |

## 下载安装

1. 从 [releases](https://gitee.com/XmirrorSecurity/OpenSCA-cli/releases) 下载对应系统架构的可执行文件压缩包

2. 或者下载源码编译(需要 `go 1.18` 及以上版本)

   ```shell
   git clone https://gitee.com/XmirrorSecurity/OpenSCA-cli.git opensca
   cd opensca
   go work init cli analyzer util
   go build -o opensca-cli cli/main.go
   ```

   默认生成当前系统架构的程序，如需生成其他系统架构可配置环境变量后编译

   - 禁用`CGO_ENABLED`
     `CGO_ENABLED=0`
   - 指定操作系统
     `GOOS=${OS} \\ darwin,freebsd,liunx,windows`
   - 指定体系架构
     `GOARCH=${arch} \\ 386,amd64,arm`

## 使用样例

仅检测组件信息

```shell
opensca-cli -path ${project_path}
```

使用本地漏洞库

```shell
opensca-cli -db db.json -path ${project_path}
```

仅使用云漏洞库服务

```shell
opensca-cli -url ${url} -token ${token} -path ${project_path}
```

使用v2.0.0及以上版本连接SaaS服务，检测漏洞信息、获取资产清单、生成数据看板并进行项目管理

```shell
opensca-cli -url ${url} -token ${token} -v2 -path ${project_path}
```

## 参数说明

**可在配置文件中配置参数，也可在命令行输入参数，两者冲突时优先使用输入参数**

| 参数       | 类型     | 描述                                                                                                                                              | 使用样例                          |
| ---------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| `config`   | `string` | 指定配置文件路径，程序启动时将配置文件中的参数作为启动参数，配置参数与命令行输入参数冲突时优先使用输入参数                                        | `-config config.json`             |
| `path`     | `string` | 指定要检测的文件或目录路径                                                                                                                        | `-path ./foo`                     |
| `url`      | `string` | 从云漏洞库查询漏洞，指定要连接云服务的地址，与 `token` 参数一起使用                                                                               | `-url https://opensca.xmirror.cn` |
| `token`    | `string` | 云服务验证 `token`，需要在云服务平台申请，与 `url` 参数一起使用                                                                                   | `-token xxxxxxx`                  |
| `v2`       | `bool`   | 是否使用新的v2版本的云服务接口                                                                                                                    | `-v2`                             |
| `cache`    | `bool`   | 建议开启，缓存下载的文件(例如 `.pom` 文件)，重复检测相同组件时会节省时间，下载的文件会保存到工具所在目录的.cache 目录下                           | `-cache`                          |
| `vuln`     | `bool`   | 结果仅保留有漏洞信息的组件，使用该参数将不会保留组件层级结构                                                                                      | `-vuln`                           |
| `out`      | `string` | 将检测结果保存到指定文件，根据后缀生成不同格式的文件，默认为 `json` 格式；支持以`spdx`格式展示`sbom`清单，只需更换相应输出文件后缀即可            | `-out output.json`                |
| `db`       | `string` | 指定本地漏洞库文件，希望使用自己漏洞库时可用，漏洞库文件为 `json` 格式，具体格式会在之后给出;若同时使用云端漏洞库与本地漏洞库，漏洞查询结果取并集 | `-db db.json`                     |
| `progress` | `bool`   | 显示进度条                                                                                                                                        | `-progress`                       |
| `dedup`    | `bool`   | 相同组件去重                                                                                                                                      | `-dedup`                          |
| `version`  | `bool`   | 显示客户端版本                                                                                                                                    | `-version`                        |

---

如果要配置maven私服库，需要在配置文件里进行配置，格式如下：

```json
{
    "maven": [
        {
            "repo": "url",
            "user": "user",
            "password": "password"
        }
    ]
}
```

### 漏洞库文件格式

```json
[
  {
    "vendor": "org.apache.logging.log4j",
    "product": "log4j-core",
    "version": "[2.0-beta9,2.12.2)||[2.13.0,2.15.0)",
    "language": "java",
    "name": "Apache Log4j2 远程代码执行漏洞",
    "id": "XMIRROR-2021-44228",
    "cve_id": "CVE-2021-44228",
    "cnnvd_id": "CNNVD-202112-799",
    "cnvd_id": "CNVD-2021-95914",
    "cwe_id": "CWE-502,CWE-400,CWE-20",
    "description": "Apache Log4j是美国阿帕奇（Apache）基金会的一款基于Java的开源日志记录工具。\r\nApache Log4J 存在代码问题漏洞，攻击者可设计一个数据请求发送给使用 Apache Log4j工具的服务器，当该请求被打印成日志时就会触发远程代码执行。",
    "description_en": "Apache Log4j2 2.0-beta9 through 2.12.1 and 2.13.0 through 2.15.0 JNDI features used in configuration, log messages, and parameters do not protect against attacker controlled LDAP and other JNDI related endpoints. An attacker who can control log messages or log message parameters can execute arbitrary code loaded from LDAP servers when message lookup substitution is enabled. From log4j 2.15.0, this behavior has been disabled by default. From version 2.16.0, this functionality has been completely removed. Note that this vulnerability is specific to log4j-core and does not affect log4net, log4cxx, or other Apache Logging Services projects.",
    "suggestion": "2.12.1及以下版本可以更新到2.12.2，其他建议更新至2.15.0或更高版本，漏洞详情可参考：https://github.com/apache/logging-log4j2/pull/608 \r\n1、临时解决方案，适用于2.10及以上版本：\r\n\t（1）设置jvm参数：“-Dlog4j2.formatMsgNoLookups=true”；\r\n\t（2）设置参数：“log4j2.formatMsgNoLookups=True”；",
    "attack_type": "远程",
    "release_date": "2021-12-10",
    "security_level_id": 1,
    "exploit_level_id": 1
  },
  {}
]
```

#### 漏洞库字段说明

| 字段                | 描述                              | 是否必填 |
| :------------------ | :-------------------------------- | :------- |
| `vendor`            | 组件厂商                          | 否       |
| `product`           | 组件名                            | 是       |
| `version`           | 漏洞影响版本                      | 是       |
| `language`          | 组件语言                          | 是       |
| `name`              | 漏洞名                            | 否       |
| `id`                | 自定义编号                        | 是       |
| `cve_id`            | cve 编号                          | 否       |
| `cnnvd_id`          | cnnvd 编号                        | 否       |
| `cnvd_id`           | cnvd 编号                         | 否       |
| `cwe_id`            | cwe 编号                          | 否       |
| `description`       | 漏洞描述                          | 否       |
| `description_en`    | 漏洞英文描述                      | 否       |
| `suggestion`        | 漏洞修复建议                      | 否       |
| `attack_type`       | 攻击方式                          | 否       |
| `release_date`      | 漏洞发布日期                      | 否       |
| `security_level_id` | 漏洞风险评级(1~4 风险程度递减)    | 否       |
| `exploit_level_id`  | 漏洞利用评级(0:不可利用,1:可利用) | 否       |

## 贡献者

- 张涛
- 张弛
- 陈钟
- 刘恩炙
- 宁戈

## 问题反馈&联系我们

微信技术交流群：（扫码添加小助手邀您入群）

<img src="https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/wechat.jpg" />

QQ技术交流群：832039395

官方邮箱：opensca@anpro-tech.com


## 向我们贡献

**OpenSCA** 是一款开源的软件成分分析工具，项目成员期待您的贡献。

如果您对此有兴趣，请参考我们的[贡献指南](./docs/贡献指南（中文版）v1.0.md)。