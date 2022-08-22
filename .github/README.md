

<p align="center">
	<img alt="logo" src="../logo.svg">
</p>
<h1 align="center" style="margin: 30px 0 30px; font-weight: bold;">OpenSCA-Cli</h1>
<p align="center">
	<a href="https://github.com/XmirrorSecurity/OpenSCA-cli/blob/master/LICENSE"><img src="https://img.shields.io/github/license/XmirrorSecurity/OpenSCA-cli?style=flat-square"></a>
	<a href="https://github.com/XmirrorSecurity/OpenSCA-cli/releases"><img src="https://img.shields.io/github/v/release/XmirrorSecurity/OpenSCA-cli?style=flat-square"></a>
</p>



## Introduction

OpenSCA is intended for scanning the third-party component dependencies and vulnerabilities.

------

## Detection Ability

OpenSCA is now capable of parsing configuration files in the listed programming languages and correspondent package managers. The project team is now dedicated to introducing more languages and enriching the parsing of relevant configuration files gradually.

| LANGUAGE     | PACKAGE MANAGER | FILE                                                         |
| ------------ | --------------- | ------------------------------------------------------------ |
| `Java`       | `Maven`         | `pom.xml`                                                    |
| `Java`       | `Gradle`        | `.gradle` `.gradle.kts`                                      |
| `JavaScript` | `Npm`           | `package-lock.json` `package.json` `yarn.lock`               |
| `PHP`        | `Composer`      | `composer.json`  `composer.lock`                             |
| `Ruby`       | `gem`           | `gemfile.lock`                                               |
| `Golang`     | `gomod`         | `go.mod` `go.sum`                                            |
| `Rust`       | `cargo`         | `Cargo.lock`                                                 |
| `Erlang`     | `Rebar`         | `rebar.lock`                                                 |
| `Python`     | `Pip`           | `Pipfile` `Pipfile.lock` `setup.py``requirements.txt``requirements.in`(For the latter two, you need to install pipenv in advance) |

## Download and Deployment

1. Download the appropriate executable file according to your system architecture from [release](https://github.com/XmirrorSecurity/OpenSCA-cli/releases).  

2. Or download the source code and compile (go 1.18 and above is needed)

   ```
   git clone https://github.com/XmirrorSecurity/OpenSCA-cli.git opensca
   cd opensca
   go work init cli analyzer util
   go build -o opensca-cli cli
   ```

   The default option is to generate the program of the current system architecture. If you want to try it for other system architectures, you can set the following environment variables before compiling.

   - Disable `CGO_ENABLED` `CGO_ENABLED=0`
   - Set the operating system `GOOS=${OS} \\ darwin,freebsd,liunx,windows`
   - Set the architecture `GOARCH=${arch} \\ 386,amd64,arm`

## Samples

For detecting the component information only:

```
opensca-cli -path ${project_path}
```

For connecting to the cloud platform:

```
opensca-cli -url ${url} -token ${token} -path ${project_path}
```

Or for using the local vulnerability database:

```
opensca-cli -db db.json -path ${project_path}
```

## Parameters

**You can either configure the parameters in configuration files or input the parameters in the command-line. When the two conflict with each other, the input parameters will be prioritized.**

| PARAMETER  | TYPE     | DESCRIPTION                                                  | SAMPLE                            |
| ---------- | -------- | ------------------------------------------------------------ | --------------------------------- |
| `config`   | `string` | Set the configuration file path, when the program runs, the parameter of the configuration file will be used as the startup parameters. If the configuration parameter conflicts with the command-line input parameter, the latter will be taken. | `-config config.json`             |
| `path`     | `string` | Set the file or directory path to be detected.               | `-path ./foo`                     |
| `url`      | `string` | Check the vulnerabilities from the cloud vulnerability database, set the address of the cloud service. It needs to be used with the `token` parameter. | `-url https://opensca.xmirror.cn` |
| `token`    | `string` | Cloud service verification. You have to apply for it on the cloud service platform and use it with the `url` parameter. | `-token xxxxxxx`                  |
| `cache`    | `bool`   | This option is recommended. It can cache the downloaded files, for example, the `.pom` file, and save your time when detecting the same component next time. The downloaded files are saved in `.cache` under the same directory as opensca-cli. | `-cache`                          |
| `vuln`     | `bool`   | Show the vulnerabilities info only. Using this parameter, the component hierarchical architecture will **NOT** be included in the result. | `-vuln`                           |
| `out`      | `string` | Set the output file. The result defaults to json format.Support the output of SBOM list in spdx format. | `-out output.json`                |
| `db`       | `string` | Set the local vulnerability database file. It helps when you prefer to use your own vulnerability database. The format of the vulnerability database is shown below. If the cloud and local vulnerability databases are both set, the result of detection will merge both. | `-db db.json`                     |
| `progress` | `bool`   | Show the progress bar.                                       | `-progress`                       |
| `dedup`    | `bool`   | Same result deduplication                                    | `-dedup`                          |

------

### The Format of the Vulnerability Database File

```
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

#### Explanations of Vulnerability Database Fields

| FIELD               | DESCRIPTION                                                       | REQUIRED OR NOT |
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
| `description`       | the description of the vulnerability                              | N               |
| `description_en`    | the description of the vulnerability in English                   | N               |
| `suggestion`        | the suggestion for fixing the vulnerability                       | N               |
| `attack_type`       | the type of attack                                                | N               |
| `release_date`      | the release date of the vulnerability                             | N               |
| `security_level_id` | the security level of the vulnerability (diminishing from 1 to 4) | N               |
| `exploit_level_id`  | the exploit level of the vulnerability (0-N/A 1-Available)        | N               |

## Contributing 

OpenSCA is an open source project, we appreciate your help!

To contribute, please read our [Contributing Guideline](../docs/Contributing%20Guideline-en%20v1.0.md). 



*For the Chinese version of this document, please check [README](../README.md).
