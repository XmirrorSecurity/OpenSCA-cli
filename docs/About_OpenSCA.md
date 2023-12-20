[Go Back](/docs/README.md) | [中文](./About_OpenSCA-zh_CN.md)

# About 
> Manage Open Source Risks Through an Open Source Solution

OpenSCA is the open source realization of SCA (Software Composition Analysis) technology. As the open source version of Xmirror SCA, it has been endowed with the core abilities of mixed-source application security detection. Aiming at guarding open source security, it is competent to dig out the hiding vulnerabilities and compliance risks in all components by dependency analysis, characteristic analysis, reference identification and compliance analysis.

Unlike traditional commercial SCA tools, OpenSCA has offered an open source solution to the management of open source risks which is full of potential. Being both complete in ability and easy to use, it supports various scenarios including online/offline, IDE/CMD/SaaS, etc. while allows customized configuration such as local vulnerability databse and private repos. Generally speaking, OpenSCA is intended for outputting transparent component assets & risk list for companies, organizations and individual developers in a flexible way.

Based on OpenSCA, we've built up a global community covering industries of telecom, internet, IoV, finance, energy and so on. We sincerely hope that our project can be a stage for communication and innovation of open source stakeholders.

# Language & Package Manager

| Language | Package Manager | File |
| :--:| :--: | :-- |
| Java | Maven | `pom.xml` |
| | Gradle | `.gradle`, `.gradle.kts` |
| JavaScripts | NPM | `package-lock.json`, `package.json`, `yarn.lock` |
| PHP | Composer | `composer.json`, `composer.lock` |
| Ruby | gem | `gemfile.lock` |
| Golang | Go mod | `go.mod`, `go.sum` |
| Python | Pip | `Pipfile`, `Pipfile.lock`, `setup.py`, `requirements.txt`(pipenv & internet needed), `requirements.in`(pipenv & internet needed) |
| Rust | cargo | `Cargo.lock` |
| Erlang | Rebar | `rebar.lock` |

# Work Flow

![DetectionProcess](/resources/DetectionProcess.png)
