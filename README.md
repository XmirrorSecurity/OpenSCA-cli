<div align="center">
	<img alt="logo" src="/resources/logo.svg">
  <h2>用开源的方式做开源风险治理</h2>

[![Release](https://img.shields.io/github/v/release/XmirrorSecurity/OpenSCA-cli)](https://github.com/XmirrorSecurity/OpenSCA-cli/releases)
[![GitHub all releases](https://img.shields.io/github/downloads/XmirrorSecurity/OpenSCA-cli/total)](https://github.com/XmirrorSecurity/OpenSCA-cli/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/opensca/opensca-cli)](https://hub.docker.com/r/opensca/opensca-cli)
[![Jetbrains Plugin](https://img.shields.io/jetbrains/plugin/v/18246)](https://plugins.jetbrains.com/plugin/18246-opensca-xcheck)
[![VSCode Plugin](https://vsmarketplacebadges.dev/version/xmirror.opensca.svg)](https://marketplace.visualstudio.com/items?itemName=xmirror.opensca)
[![LICENSE](https://img.shields.io/github/license/XmirrorSecurity/OpenSCA-cli)](https://github.com/XmirrorSecurity/OpenSCA-cli/blob/master/LICENSE)
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/xmirrorsecurity/opensca-cli)](/go.mod)

中文 | [English](./.github/README.md)
</div>

<!-- TOC -->
- [项目介绍](#项目介绍)
- [检测能力](#检测能力)
- [下载安装](#下载安装)
  - [方式 1: 从 Releases 下载](#方式-1-从-releases-下载)
  - [方式 2: 一键安装脚本](#方式-2-一键安装脚本)
  - [方式 3: 使用包管理器(Homebrew)](#方式-3-使用包管理器homebrew)
  - [方式 4: 从源码构建](#方式-4-从源码构建)
- [使用说明](#使用说明)
  - [参数说明](#参数说明)
  - [报告格式](#报告格式)
  - [使用样例](#使用样例)
  - [漏洞库文件格式](#漏洞库文件格式)
  - [漏洞库字段说明](#漏洞库字段说明)
  - [漏洞库配置示例](#漏洞库配置示例)
- [常见问题](#常见问题)
  - [使用OpenSCA需要配置环境变量吗？](#使用opensca需要配置环境变量吗)
  - [OpenSCA目前支持哪些漏洞库呢？](#opensca目前支持哪些漏洞库呢)
  - [使用OpenSCA检测时，检测速度与哪些因素有关？](#使用opensca检测时检测速度与哪些因素有关)
- [问题反馈\&联系我们](#问题反馈联系我们)
- [贡献者](#贡献者)
- [向我们贡献](#向我们贡献)
<!-- TOC -->

## 项目介绍

**OpenSCA** 用来扫描项目的第三方组件依赖及漏洞信息。

官网：[https://opensca.xmirror.cn](https://opensca.xmirror.cn)

欢迎点亮**star**，鼓励下项目组的小伙伴们~

---

## 检测能力

`OpenSCA`现已支持以下编程语言相关的配置文件解析及对应的包管理器，后续会逐步支持更多的编程语言，丰富相关配置文件的解析。

| 支持语言     | 包管理器   | 解析文件                                                                 |
| ------------ | ---------- | ------------------------------------------------------------------------ |
| `Java`       | `Maven`    | `pom.xml`                                                                |
| `Java`       | `Gradle`   | `.gradle` `.gradle.kts`                                                  |
| `JavaScript` | `Npm`      | `package-lock.json` `package.json` `yarn.lock`                           |
| `PHP`        | `Composer` | `composer.json` `composer.lock`                                          |
| `Ruby`       | `gem`      | `gemfile.lock`                                                           |
| `Golang`     | `gomod`    | `go.mod` `go.sum` `Gopkg.toml` `Gopkg.lock`                              |
| `Rust`       | `cargo`    | `Cargo.lock`                                                             |
| `Erlang`     | `Rebar`    | `rebar.lock`                                                             |
| `Python`     | `Pip`      | `Pipfile` `Pipfile.lock` `setup.py` `requirements.txt` `requirements.in` |

## 下载安装

OpenSCA-cli 支持 Windows、Linux、MacOS 等操作系统，支持 x86_64 和 arm64 架构。目前提供了以下几种安装方式：

### 方式 1: 从 Releases 下载

1. 从 [github](https://github.com/XmirrorSecurity/OpenSCA-cli/releases) 或 [gitee](https://gitee.com/XmirrorSecurity/OpenSCA-cli/releases) 或 [gitcode](https://gitcode.com/XmirrorSecurity/OpenSCA-cli/releases) 下载对应系统架构的可执行文件压缩包
2. 解压，并运行 `opensca-cli` 可执行文件即可。

### 方式 2: 一键安装脚本

- Mac/Linux 用户
    ```shell
    curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh

    # 如果遇到网络问题，可尝试以下命令
    curl -sSL https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/scripts/install.sh | sh -s -- gitee
    ```
- For Windows Users(need PowerShell)
    ```powershell
    iex "&{$(irm https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.ps1)}"

    # 如遇到网络问题，可尝试以下命令
    iex "&{$(irm https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/scripts/install.ps1)} gitee"
    ```

### 方式 3: 使用包管理器(Homebrew)

```
brew install opensca-cli
```

### 方式 4: 从源码构建

克隆并源码编译(需要 `go 1.18` 及以上版本)

```shell
# github linux/mac
git clone https://github.com/XmirrorSecurity/OpenSCA-cli.git opensca && cd opensca && go build
# gitee linux/mac
git clone https://gitee.com/XmirrorSecurity/OpenSCA-cli.git opensca && cd opensca && go build
# gitcode linux/mac
git clone https://gitcode.com/XmirrorSecurity/OpenSCA-cli.git opensca && cd opensca && go build
# github windows
git clone https://github.com/XmirrorSecurity/OpenSCA-cli.git opensca ; cd opensca ; go build
# gitee windows
git clone https://gitee.com/XmirrorSecurity/OpenSCA-cli.git opensca ; cd opensca ; go build
# gitcode windows
git clone https://gitcode.com/XmirrorSecurity/OpenSCA-cli.git opensca ; cd opensca ; go build
```

默认生成当前系统架构的程序，如需生成其他系统架构可配置环境变量后编译

- 禁用`CGO_ENABLED`
  `CGO_ENABLED=0`
- 指定操作系统
  `GOOS=${OS} \\ darwin,liunx,windows`
- 指定体系架构
  `GOARCH=${arch} \\ amd64,arm64`

## 使用说明

### 参数说明

| 参数     | 类型     | 描述             | 使用样例                 |
| -------- | -------- | ---------------- | ------------------------ |
| `config` | `string` | 指定配置文件路径 | `-config config.json`    |
| `path`   | `string` | 指定检测项目路径 | `-path ./foo`            |
| `out`    | `string` | 根据后缀生成报告 | `-out out.json,out.html` |
| `log`    | `string` | 指定日志文件路径 | `-log my_log.txt`        |
| `token`  | `string` | 云端服务`token`  | `-token xxx`             |
| `proj`   | `string` | saas项目`token`  | `-proj xxx`              |

完整的检测参数需在配置文件中配置
（*v3.0.0开始url参数不再通过命令行指定，默认为OpenSCA云漏洞库服务`https://opensca.xmirror.cn/`，也可通过配置文件指定其他数据格式相符的云漏洞库；使用过往版本可在命令行或配置文件指定url参数。）

配置字段及说明详见[`config.json`](./config.json)

v3.0.2开始，OpenSCA-cli可以通过proj参数向OpenSCA SaaS同步检出结果，方便资产及风险的全局管理。

配置文件与命令行参数冲突时优先使用命令行输入参数

指定了配置文件路径但目标位置不存在文件时会在目标位置生成默认配置文件

未指定配置文件路径会依次尝试访问以下位置:

1. 工作目录下的`config.json`
2. 用户目录下的`opensca_config.json`
3. `opensca-cli`目录下的`config.json`

### 报告格式

`out` 参数支持范围如下：

| 类型     | 文件格式 | 识别的文件后缀                   |
| -------- | -------- | -------------------------------- |
| 检测报告 | `json`   | `.json`                          |
|          | `xml`    | `.xml`                           |
|          | `html`   | `.html`                          |
|          | `sqlite` | `.sqlite`                        |
|          | `csv`    | `.csv`                           |
|          | `sarif`  | `.sarif`                         |
| SBOM清单 | `spdx`   | `.spdx` `.spdx.json` `.spdx.xml` |
|          | `cdx`    | `.cdx.json` `.cdx.xml`           |
|          | `swid`   | `.swid.json` `.swid.xml`         |
|          | `dsdx`   | `.dsdx` `.dsdx.json` `.dsdx.xml` |

### 使用样例

```shell
# 使用opensca-cli检测
opensca-cli -path ${project_path} -config ${config_path} -out ${filename}.${suffix} -token ${token}

# 写好配置文件后也可以直接执行opensca-cli
opensca-cli
```

```shell
# 检测当前目录的依赖信息
docker run -ti --rm -v ${PWD}:/src opensca/opensca-cli

# 使用云端漏洞数据库:
docker run -ti --rm -v ${PWD}:/src opensca/opensca-cli -token ${put_your_token_here}
```

如需在`docker`容器中使用配置文件，将`config.json`放到`src`挂载目录即可。也可以使用`-config`指定其他容器内路径。
不同终端挂载当前目录的写法不同，常见的几种终端写法如下：
| terminal     | pwd                   |
| ------------ | --------------------- |
| `bash`       | `$(pwd)`              |
| `zsh`        | `${PWD}`              |
| `cmd`        | `%cd%`                |
| `powershell` | `(Get-Location).Path` |

更多信息请参考 [Docker Hub 主页](https://hub.docker.com/r/opensca/opensca-cli)

---

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
  }
]
```

### 漏洞库字段说明

| 字段                | 描述                                     | 是否必填 |
| :------------------ | :--------------------------------------- | :------- |
| `vendor`            | 组件厂商                                 | 否       |
| `product`           | 组件名                                   | 是       |
| `version`           | 漏洞影响版本(必须为范围，不能填单个版本) | 是       |
| `language`          | 组件语言                                 | 是       |
| `name`              | 漏洞名                                   | 否       |
| `id`                | 自定义编号                               | 是       |
| `cve_id`            | cve 编号                                 | 否       |
| `cnnvd_id`          | cnnvd 编号                               | 否       |
| `cnvd_id`           | cnvd 编号                                | 否       |
| `cwe_id`            | cwe 编号                                 | 否       |
| `description`       | 漏洞描述                                 | 否       |
| `description_en`    | 漏洞英文描述                             | 否       |
| `suggestion`        | 漏洞修复建议                             | 否       |
| `attack_type`       | 攻击方式                                 | 否       |
| `release_date`      | 漏洞发布日期                             | 否       |
| `security_level_id` | 漏洞风险评级(1~4 风险程度递减)           | 否       |
| `exploit_level_id`  | 漏洞利用评级(0:不可利用,1:可利用)        | 否       |

本地漏洞库中`language`字段设定值包含`java、javascript、golang、rust、php、ruby、python`

`version`范围描述：

| 符号          | 描述(`x`为检出的组件版本)        |
| ------------- | -------------------------------- |
| `[a,b]`       | `a<=x<=b`                        |
| `(a,b)`       | `a<x<b`                          |
| `[a,b)`       | `a<=x<b`                         |
| `(a,b]`       | `a<x<=b`                         |
| `(0,b)`       | `x<b`                            |
| `(a,)`        | `x>a`                            |
| `{a,b,c,...}` | `x=a` 或 `x=b` 或 `x=c` 或 `...` |

同时位于多个范围需要用`||`连接，例如：
`[a,b)||(b,c]`代表`a<=x<b`或`b<x<=c`，即`a<=x<=c`且`x!=b`
也可以区间和集合混用：
`(0,b)||{c,d}||[e,)`代表`x<b`或`x=c`或`x=d`或`x>=e`

### 漏洞库配置示例

```json
{
  "origin":{
    "json":"db.json",
    "mysql":{
      "dsn":"user:password@tcp(ip:port)/dbname",
      "table":"table_name"
    },
    "sqlite":{
      "dsn":"sqlite.db",
      "table":"table_name"
    }
  }
}
```

## 常见问题

### 使用OpenSCA需要配置环境变量吗？

不需要。解压后直接在命令行或终端工具中执行对应命令即可开始检测。

### OpenSCA目前支持哪些漏洞库呢？

OpenSCA支持自主配置本地漏洞库，需要按照[漏洞库文件格式](https://opensca.xmirror.cn/docs/v1/cli.html#漏洞库文件格式)配置。

同时OpenSCA提供云漏洞库服务，兼容NVD、CNVD、CNNVD等官方漏洞库。

### 使用OpenSCA检测时，检测速度与哪些因素有关？

检测速度与压缩包大小、网络状况和检测语言有关，通常情况下会在几秒到几分钟。

v1.0.11开始在默认逻辑中新增了阿里云镜像库作为maven官方库的备用，解决了官方库连接受限导致的检测速度过慢问题。

v1.0.10及更低版本使用时如遇检测速度异常慢、日志文件中有maven连接失败报错，v1.0.6-v1.0.10可在配置文件config.json中将“maven”字段作如下设置：

```json
{
    "maven": [
        {
            "repo": "https://maven.aliyun.com/repository/public",
            "user": "",
            "password": ""
        }
    ]
}
```

设置完毕后，确保配置文件和opensca-cli在同一目录下，执行opensca-cli检测命令加上-config congif.json即可，示例：

```shell
opensca-cli -token {token} -path {path} -out output.html -config config.json
```

v1.0.5及更低版本需要自行修改源码配置镜像库地址，建议升级到更高版本。

**更多常见问题**，参见[常见问题](https://opensca.xmirror.cn/docs/v1/FAQ.html)。

## 问题反馈&联系我们

如果您在使用中遇到问题，欢迎向我们提交ISSUE。

也可添加下方微信：

![二维码](/resources/wechat.png)

QQ技术交流群：832039395

官方邮箱：<opensca@anpro-tech.com>

## 贡献者

- 张涛
- 张弛
- 陈钟
- 刘恩炙
- 宁戈

## 向我们贡献

**OpenSCA** 是一款开源的软件成分分析工具，项目成员期待您的贡献。

如果您对此有兴趣，请参考我们的[贡献指南](./docs/Contributing_Guideline-v1.0-zh_CN.md)。
