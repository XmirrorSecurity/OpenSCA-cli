---
creator: Cyber Chen
editor: Cyber Chen
modification time: 2023-11-29 11:22:00
---

[返回目录](/docs/README-zh-CN.md) | [English](./About_OpenSCA.md)

# 关于 OpenSCA
> 用开源的方式做开源风险治理

OpenSCA 是 SCA 技术原理的开源实现。作为[悬镜安全](https://www.xmirror.cn)旗下[源鉴SCA开源威胁管控产品](https://oss.xmirror.cn/)的开源版本，OpenSCA继承了源鉴SCA的多源SCA开源应用安全缺陷检测等核心能力，通过软件成分分析、依赖分析、特征分析、引用识别、合规分析等方法，深度挖掘组件中潜藏的各类安全漏洞及开源协议风险，保障应用开源组件引入的安全。

不同于传统企业版SCA工具，OpenSCA为治理开源风险提供了充满可能性的开源解决方案。它轻量易用、能力完整，支持漏洞库、私服库等自主配置，覆盖IDE/命令行/云平台、离线/在线等多种使用场景，可灵活地接入开发流程，为企业、组织及个人用户输出透明化的组件资产及风险清单。

围绕OpenSCA，我们搭建起了聚集上万开源项目维护者和使用者的全球极客开源数字供应链安全社区，社区涵盖信息通信、泛互联网、车联网、金融、能源等众多行业用户，为万千中国数字安全实践者们构筑起交流的平台与创新的基地。

# 支持语言 & 包管理器

| 语言 | 包管理器 | 特征文件 |
| :--:| :--: | :-- |
| Java | Maven | `pom.xml` |
| | Gradle | `.gradle`, `.gradle.kts` |
| JavaScripts | NPM | `package-lock.json`, `package.json`, `yarn.lock` |
| PHP | Composer | `composer.json`, `composer.lock` |
| Ruby | gem | `gemfile.lock` |
| Golang | Go mod | `go.mod`, `go.sum` |
| Python | Pip | `Pipfile`, `Pipfile.lock`, `setup.py`, `requirements.txt`(依赖 pipenv, 需联网), `requirements.in`(依赖 pipenv, 需联网) |
| Rust | cargo | `Cargo.lock` |
| Erlang | Rebar | `rebar.lock` |

# 检测流程



