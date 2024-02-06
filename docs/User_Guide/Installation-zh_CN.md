[返回目录](/docs/README-zh-CN.md) | [English](./Installation.md)

# 一键安装

- Windows(需要 PowerShell)
  ```powershell
  iex "&{$(irm https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.ps1)}"

  # 如果在下载中遇到网络问题，可尝试使用以下命令
  iex "&{$(irm https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/scripts/install.ps1)} gitee"
  ```
- Linux/MacOS
  ```shell
  curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh

  # 如果在下载中遇到网络问题，可尝试使用以下命令
  curl -sSL https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/scripts/install.sh | sh -s -- gitee
  ```

# 使用包管理器安装

- Windows 通过 [Winget](https://github.com/microsoft/winget-cli) 安装
  ```shell
  winget install opensca-cli
  ```
- Windows 通过 [Scoop](https://scoop.sh/) 安装
  ```shell
  scoop bucket add extras
  scoop install extras/opensca-cli
  ```
- MacOS/Linux 通过 [Homebrew](https://brew.sh/) 安装 
  ```shell
  brew install opensca-cli
  ```

# 手动安装

从 [Github Release](https://github.com/XmirrorSecurity/OpenSCA-cli/releases) 或 [Gitee Release](https://gitee.com/XmirrorSecurity/OpenSCA-cli/releases/) 下载对应系统和处理器架构的压缩包，解压到任意目录即可使用。

# 从源码构建

依赖 Go 1.18+

- 克隆此仓库
  ```shell
  git clone https://github.com/XmirrorSecurity/OpenSCA-cli.git opensca && cd opensca
  ```
- 编译
  ```shell
  go build
  ```
  使用此命令会生成当前系统和处理器架构的可执行文件，如需生成其他系统架构可配置环境变量后编译
  - 禁用CGO_ENABLED CGO_ENABLED=0
  - 指定操作系统 GOOS=${OS} \\ darwin,liunx,windows
  - 指定体系架构 GOARCH=${arch} \\ amd64,arm64
