---
title: Quick Start
author: Cyber Chen
date: 2023-11-30T11:01:15+08:00
---

[返回目录](/docs/README-zh-CN.md) | [English](./Quick_Start.md)

# 快速开始

## 视频教程

<video width="320" height="240" controls="controls" poster="https://opensca.xmirror.cn/docs/assets/img/poster_cli.d9973be2.png" style="width: 100%; max-height: 500px; height: auto;" jm_neat="328132609"><source src="https://opensca.xmirror.cn/docs/assets/media/cli.1bed8c1c.mp4" type="video/mp4">
您的浏览器不支持 video 标签。
</video>

## 传统方式

### 下载安装

- 方式一：使用一键安装脚本
  - Mac/Linux 用户可通过以下命令下载并安装
    ```shell
    curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh

    # 如果在下载中遇到网络问题，可尝试使用以下命令
    curl -sSL https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/scripts/install.sh | sh -s -- gitee
    ```
  - Windows 用户可通过以下命令下载并安装(Powershell)
    ```powershell
    iex "&{$(irm https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.ps1)}"

    # 如果在下载中遇到网络问题，可尝试使用以下命令
    iex "&{$(irm https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/scripts/install.ps1)} gitee"
    ```

- 方式二：通过包管理器安装

  - Windows [Winget](https://github.com/microsoft/winget-cli) 安装
    ```shell
    winget install opensca-cli
    ```
  - Windows [Scoop](https://scoop.sh/) 安装
    ```shell
    scoop bucket add extras
    scoop install extras/opensca-cli
    ```
  - Mac/Linux 用户可通过 [Homebrew](https://brew.sh/) 安装
    ```shell
    brew install opensca-cli
    ```

- 方式三：从 [GitHub](https://github.com/XmirrorSecurity/OpenSCA-cli/releases/latest) 或 [Gitee](https://gitee.com/XmirrorSecurity/OpenSCA-cli/releases/latest) 下载对应系统架构的可执行程序压缩包，并解压到本地任意目录下

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