---
title: Quick Start
author: Cyber Chen
date: 2023-11-30T11:01:15+08:00
---

# 快速开始

## 视频教程

<video width="320" height="240" controls="controls" poster="https://opensca.xmirror.cn/docs/assets/img/poster_cli.d9973be2.png" style="width: 100%; max-height: 500px; height: auto;" jm_neat="328132609"><source src="https://opensca.xmirror.cn/docs/assets/media/cli.1bed8c1c.mp4" type="video/mp4">
您的浏览器不支持 video 标签。
</video>

## 传统方式

### 下载安装

- 方式一：使用一键安装脚本(TODO)
- 方式二：Mac/Linux 用户可通过 `Homebrew` 下载安装
  ```shell
  brew install opensca-cli
  ```
- 方式三：从 [GitHub]() 或 [Gitee]() 下载对应系统架构的可执行程序压缩包，并解压到本地任意目录下

### 开始检测

**检测指定目录的依赖关系**

```shell
opensca-cli -path {替换为要检测的目录}
```

**检测指定目录的依赖关系，并通过云端数据库获取许可证以及漏洞信息**

> 您需要先[注册](https://opensca.xmirror.cn/register)并获取 token

```shell
opensca-cli -path {替换为要检测的目录} -token {替换为您的 token}
```

## Docker

**检测指定目录的依赖关系**

```shell
docker run -ti --rm -v {替换为要检测的目录}:/src opensca/opensca-cli:latest
```

**检测指定目录的依赖关系，并通过云端数据库获取许可证以及漏洞信息**

> 您需要先[注册](https://opensca.xmirror.cn/register)并获取 token

```shell
docker run -ti --rm -v {替换为要检测的目录}:/src opensca/opensca-cli:latest -token {替换为您的 token}
```