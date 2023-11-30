[![Release](https://img.shields.io/github/v/release/XmirrorSecurity/OpenSCA-cli)](https://github.com/XmirrorSecurity/OpenSCA-cli/releases)
[![GitHub all releases](https://img.shields.io/github/downloads/XmirrorSecurity/OpenSCA-cli/total)](https://github.com/XmirrorSecurity/OpenSCA-cli/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/opensca/opensca-cli)](https://hub.docker.com/r/opensca/opensca-cli)
[![Jetbrains Plugin](https://img.shields.io/jetbrains/plugin/v/18246)](https://plugins.jetbrains.com/plugin/18246-opensca-xcheck)
[![VSCode Plugin](https://vsmarketplacebadges.dev/version/xmirror.opensca.svg)](https://marketplace.visualstudio.com/items?itemName=xmirror.opensca)
[![LICENSE](https://img.shields.io/github/license/XmirrorSecurity/OpenSCA-cli)](https://github.com/XmirrorSecurity/OpenSCA-cli/blob/master/LICENSE)
<!--
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/xmirrorsecurity/opensca-cli)](/go.mod)
-->

<div align="center">
	<img alt="logo" src="/resources/logo.svg">
</div>

English|[中文](../README.md)

- [Introduction](#introduction)
- [Detection Ability](#detection-ability)
- [Download \& Deployment](#download--deployment)
- [Use OpenSCA](#use-opensca)
  - [Parameters](#parameters)
  - [Report Formats](#report-formats)
  - [Sample](#sample)
    - [Scan \& Report via docker container](#scan--report-via-docker-container)
  - [Local Vulnerability Database](#local-vulnerability-database)
    - [The Format of the Vulnerability Database File](#the-format-of-the-vulnerability-database-file)
    - [Explanations of Vulnerability Database Fields](#explanations-of-vulnerability-database-fields)
    - [Sample of Setting the Vulnerability Database](#sample-of-setting-the-vulnerability-database)
- [FAQ](#faq)
  - [Is the environment variable needed while using OpenSCA?](#is-the-environment-variable-needed-while-using-opensca)
  - [About the vulnerability database?](#about-the-vulnerability-database)
  - [About the time cost of OpenSCA scanning?](#about-the-time-cost-of-opensca-scanning)
- [Contact Us](#contact-us)
- [Authors](#authors)
- [Contributing](#contributing)


## Introduction

OpenSCA is intended for scanning third-party dependencies, vulnerabilities and licenses.

Our website: [https://opensca.xmirror.cn](https://opensca.xmirror.cn)

Click **STAR** to leave encouragement.

------

## Detection Ability

OpenSCA is now capable of parsing configuration files in the listed programming languages and correspondent package managers. The team is now dedicated to introducing more languages and enriching the parsing of relevant configuration files gradually.

| LANGUAGE     | PACKAGE MANAGER | FILE                                                                                                                                              |
| ------------ | --------------- |---------------------------------------------------------------------------------------------------------------------------------------------------|
| `Java`       | `Maven`         | `pom.xml`                                                                                                                                         |
| `Java`       | `Gradle`        | `.gradle` `.gradle.kts`                                                                                                                           |
| `JavaScript` | `Npm`           | `package-lock.json` `package.json` `yarn.lock`                                                                                                    |
| `PHP`        | `Composer`      | `composer.json` `composer.lock`                                                                                                                   |
| `Ruby`       | `gem`           | `gemfile.lock`                                                                                                                                    |
| `Golang`     | `gomod`         | `go.mod` `go.sum` `Gopkg.toml` `Gopkg.lock`                                                                                                       |
| `Rust`       | `cargo`         | `Cargo.lock`                                                                                                                                      |
| `Erlang`     | `Rebar`         | `rebar.lock`                                                                                                                                      |
| `Python`     | `Pip`           | `Pipfile` `Pipfile.lock` `setup.py` `requirements.txt` `requirements.in`(For the latter two, pipenv environment & internet connection are needed) |

## Download & Deployment

1. Download the appropriate executable file according to your system architecture from [releases](https://github.com/XmirrorSecurity/OpenSCA-cli/releases).

2. Or download the source code and compile (`go 1.18` and above is needed)

   ```shell
   // github
   git clone https://github.com/XmirrorSecurity/OpenSCA-cli.git opensca && cd opensca
   go build
   ```
   
   ```shell
   // gitee
   git clone https://gitee.com/XmirrorSecurity/OpenSCA-cli.git opensca && cd opensca
   go build
   ```

   The default option is to generate the program of the current system architecture. If you want to try it for other system architectures, you can set the following environment variables before compiling.

   - Disable `CGO_ENABLED` `CGO_ENABLED=0`
   - Set the operating system `GOOS=${OS} \\ darwin,liunx,windows`
   - Set the architecture `GOARCH=${arch} \\ amd64,arm64`

## Use OpenSCA

### Parameters

| PARAMETER  | TYPE     | Descripation                                                                                                                                                                                                                                                                | SAMPLE                                                                                                                                                                                                                                                          |
| ---------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `config`   | `string` | Set the path of the configuration file         | `-config config.json`                                                                                                                                                                                                                                           |
| `path`     | `string` | Set the path of the target file or directory                                                                                                                                                                                                                               | `-path ./foo`                                                                                                                                                                                                                                                   |                                                                                                            
| `out`      | `string` | Save the result to the specified file whose format is defined by the suffix | `-out out.json, out.html`  |
| `log`    | `string`   | Specify the path of log file                                                                                                                                                                                                                                                  | `-log my_log.txt`                                                                                                                                                                                                                                                        |
| `token`    | `string` | Cloud service verification coming from our offical website                                                                                                                                                   | `-token xxx`                                                                                                                                                                                                                                                |

From v3.0.0, apart from these parameters available for CMD/CRT, there are also others for different requirements which have to be set in the configuration file. 

Full introduction about each parameters can be found in `config.json`

If the configuration parameter conflicts with the command-line input parameter, the latter will be taken.

When there's no configuration file in the set path, one in default settings will be generated there.

If no path of configuration file is set, the following ones will be checked:

  1. `config.json` under the working directory
  2. `opensca_config.json` under the user directory
  3. `config.json` under `opensca-cli` directory

From v3.0.0, `url` has been put in the configuration file. The default set goes to our cloud vulnerability database. Other online database in accordance with our database structure can also be set through configuration file.  

Using previous versions to connect the cloud databse will still need the setting of `url`, which could be done via both CMD and configuration file. Example: `-url https://opensca.xmirror.cn`

### Report Formats

Files supported by the `out` parameter are listed below：

| TYPE   | FORMAT | SPECIFIED SUFFIX                 | VERSION            |
| ------ | ------ | -------------------------------- | ------------------ |
| REPORT | `json` | `.json`                          | `*`                |
|        | `xml`  | `.xml`                           | `*`                |
|        | `html` | `.html`                          | `v1.0.6` and above |
|        | `sqlite` | `.sqlite`                      | `v1.0.13` and above|
|        | `csv` | `.csv`                            | `v1.0.13` and above|
| SBOM   | `spdx` | `.spdx` `.spdx.json` `.spdx.xml` | `v1.0.8` and above |
|        | `cdx`  | `.cdx.json` `.cdx.xml`           | `v1.0.11`and above |
|        | `swid` | `.swid.json` `.swid.xml`         | `v1.0.11`and above |
|        | `dsdx` | `.dsdx` `.dsdx.json` `.dsdx.xml` | `v3.0.0`and above  |

### Sample

```shell
# Use opensca-cli to scan with CMD parameters:
opensca-cli -path ${project_path} -config ${config_path} -out ${filename}.${suffix} -token ${token}

# Start scanning after setting down the configuration file:
opensca-cli
```

#### Scan & Report via Docker Container

```shell
# Detect dependencies in the current directory:
docker run -ti --rm -v $(PWD):/src opensca/opensca-cli

# Connect to the cloud vulnerability database:
docker run -ti --rm -v $(PWD):/src opensca/opensca-cli -token ${put_your_token_here}
```

You can also use the configuration file for advanced settings. Save `config.json` to the mounted directory of `src` or set other paths within the container through `-config`.

For more information, visit [Docker Hub Page](https://hub.docker.com/r/opensca/opensca-cli)


### Local Vulnerability Database

#### The Format of the Vulnerability Database File

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

#### Explanations of Vulnerability Database Fields

| FIELD               | Descripation                                                       | REQUIRED OR NOT |
| ------------------- | ----------------------------------------------------------------- | --------------- |
| `vendor`            | the manufacturer of the component                                 | N               |
| `product`           | the name of the component                                         | Y               |
| `version`           | the versions of the component affected by the vulnerability       | Y               |
| `language`          | the programming language of the component                         | Y               |
| `name`              | the name of the vulnerability                                     | N               |
| `id`                | custom identifier                                                 | Y               |
| `cve_id`            | cve identifier                                                    | N               |
| `cnnvd_id`          | cnnvd identifier                                                  | N               |
| `cnvd_id`           | cnvd identifier                                                   | N               |
| `cwe_id`            | cwe identifier                                                    | N               |
| `description`       | the descripation of the vulnerability                              | N               |
| `description_en`    | the descripation of the vulnerability in English                   | N               |
| `suggestion`        | the suggestion for fixing the vulnerability                       | N               |
| `attack_type`       | the type of attack                                                | N               |
| `release_date`      | the release date of the vulnerability                             | N               |
| `security_level_id` | the security level of the vulnerability (diminishing from 1 to 4) | N               |
| `exploit_level_id`  | the exploit level of the vulnerability (0-N/A 1-Available)        | N               |

*There are several pre-set values to the "language" field, including java, js, golang, rust, php, ruby and python. Other languages are not limited to the pre-set value.

#### Sample of Setting the Vulnerability Database

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

## FAQ

### Is the environment variable needed while using OpenSCA?

No. OpenSCA can be directly executed by the command in CLI/CRT after decompression.

### About the vulnerability database?

OpenSCA allows configuring the local vulnerability database. It has to be sorted according to *the Format of the Vulnerability Database File*.

Meanwhile, OpenSCA also offers a cloud vulnerability database covering official databases including CVE/CWE/NVD/CNVD/CNNVD.

### About the time cost of OpenSCA scanning?

It depends on the size of the package, the network condition and the language.

From v1.0.11, we add aliyun mirror database as the backup to the official maven repository to solve the lag caused by network connection.

For v1.0.10 and below, if the time is abnormally long and error information about connection failure to the maven repository gets reported in the log file, users of versions between v1.0.6 and v1.0.10 can fix the problem by setting the `maven` field in `config.json`  like below:

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

After setting, save `config.json` to the same folder of opensca-cli.exe and execute the command. Eg:

```shell
opensca-cli -url https://opensca.xmirror.cn -token {token} -path {path} -out output.html -config config.json
```

Users of v1.0.5 and below may have to modify the source code. We recommend an upgrade to higher versions.

For more other FAQs, please check [FAQs](https://opensca.xmirror.cn/docs/v1/FAQ.html).

## Contact Us

ISSUEs are warmly welcome.

Add WeChat for further consults is also an option:

![QR Code](/resources/wechat.png)

Our QQ Group: 832039395

Mailbox: opensca@anpro-tech.com

## Authors

- Tao Zhang
- Chi Zhang
- Zhong Chen
- Enzhi Liu
- Ge Ning

## Contributing

OpenSCA is an open source project, we appreciate your contribution!

To contribute, please read our [Contributing Guideline](../docs/Contributing%20Guideline-en%20v1.0.md).
