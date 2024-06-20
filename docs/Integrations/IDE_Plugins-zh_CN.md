[返回目录](/docs/README-zh-CN.md) / [English](./IDE_Plugins.md)

- [Visual Studio Code](#visual-studio-code)
  - [安装插件](#安装插件)
  - [使用插件](#使用插件)
    - [插件功能](#插件功能)
    - [插件执行流程](#插件执行流程)
    - [运行扫描](#运行扫描)
- [Jetbrains IDEs](#jetbrains-ides)
  - [安装插件](#安装插件-1)
  - [使用插件](#使用插件-1)
    - [插件功能](#插件功能-1)
    - [插件执行流程](#插件执行流程-1)
    - [运行扫描](#运行扫描-1)


# Visual Studio Code

## 安装插件

- **安装方法 一**：在 [VS Marketplace](https://marketplace.visualstudio.com/items?itemName=xmirror.opensca) 中安装（推荐）

    在VS Code中左边栏打开扩展->扩展的搜索框中输入“OpenSCA Xcheck”，点击“Install”

    <img src="https://opensca.xmirror.cn/docs/img/vscode_01.jpg" alt="xcheck_market" />

- **安装方法二**：在[OpenSCA 官网](https://opensca.xmirror.cn/pages/plug-in)下载插件安装

  - 从OpenSCA平台下载 “OpenSCA-Xcheck.vsix”；
  - 打开VS Code，依次操作：左边栏打开扩展->扩展顶栏的更多操作->“从VSIX安装”->找到并安装“OpenSCA-Xcheck.vsix”；

- **安装方法三**：[下载源码](https://github.com/XmirrorSecurity/)自行编译安装

  - 环境要求：

    - node v18及以上版本
    - 系统支持MacOS、Windows、Linux

  - 从[gitee](https://gitee.com/XmirrorSecurity/OpenSCA-VSCode-plugin)或[github](https://github.com/XmirrorSecurity/OpenSCA-VSCode-plugin/)下载源码

  * 全局安装vsce

    ```
    npm install --global @vscode/vsce
    ```

  * 执行打包命令

    ```
    vsce package
    ```

## 使用插件

### 插件功能

- 开始检测：点击操作栏的“Run”，开始检测当前项目内的组件漏洞风险情况；
- 停止检测：点击操作栏的“Stop”，停止检测当前项目内的组件漏洞风险情况；
- 清除检测结果：点击操作栏的“Clean”，清除当前项目的检测结果；
- 连接配置：点击操作栏的“Test”按钮，配置平台Url及Token信息，点击“测试连接”按钮可测试连接配置是否正确，连接成功后就可以开始检测啦；
- 设置：点击操作栏的“Setting”，查看OpenSCA Xcheck相关设置信息。
- 使用说明：点击操作栏的“Instructions”，查看OpenSCA Xcheck相关使用说明。
- 查看更多：点击操作栏的“See more”，跳转到[opensca.xmirror.cn](https://opensca.xmirror.cn)查看OpenSCA Xcheck 更多相关信息。

<img src="https://opensca.xmirror.cn/docs/img/vscode_02.jpg" alt="xcheck_function" />

### 插件执行流程

<img src="https://opensca.xmirror.cn/docs/assets/img/xcheck_process.7083b869.jpg" alt="xcheck流程图"  />

### 运行扫描

点击OpenSCA Xcheck可打开OpenSCA窗口。首先在配置界面中配置服务器参数（参考：插件功能-设置），然后在OpenSCA窗口中点击“Run”（参考：插件功能-开始检测）

# Jetbrains IDEs

## 安装插件

- **安装方法一**：从 [Jetbrains 插件市场](https://plugins.jetbrains.com/plugin/18246-opensca-xcheck) 中安装（推荐）

    以IntelliJ IDEA为例：在IDE中依次点击“File|Settings|Plugins|Marketplace”，在搜索框中输入“OpenSCA Xcheck”，点击“Install”
    
    ![xcheck_market](/resources/xcheck_marketplace.jpg)

- **安装方法二**：在[OpenSCA平台](https://opensca.xmirror.cn/pages/plug-in )下载插件安装

    以IntelliJ IDEA为例：将下载下来的插件安装包拖入适配的IDE中即可

- **安装方法三**：[下载源码](https://github.com/XmirrorSecurity/OpenSCA-intellij-plugin )自行编译安装

    使用IntelliJ IDEA打开下载到本地的源码，需要配置运行环境：`jDK11`，待Gradle导入依赖和插件，在Gradle中执行`intellij`插件的`buildPlugin`任务，构建的安装包存放于当前项目下*build/distributions*目录下，将此目录下的安装包拖入当前IDE中即可

## 使用插件

### 插件功能

- 配置：点击File|Settings|Other Settings|OpenSCA Setting或点击OpenSCA窗口中的`Setting`按钮，在配置界面中配置连接服务器Url和Token
- 测试连接：在OpenSCA配置界面中，配置服务器Url和Token之后点击`测试连接`按钮可验证Url和Token是否有效
- 运行：点击OpenSCA窗口中的`Run`按钮，可对当前项目进行代码评估
- 停止：如果正在对当前项目代码评估，那么`Stop`按钮是可用的，点击Stop按钮可结束当前评估任务
- 清除：如果OpenSCA窗口中的Xcheck子窗口已有评估结果，点击`Clean`按钮可清除Xcheck子窗口中所有结果
![xcheck_function](/resources/xcheck_function.jpg)

### 插件执行流程

![xcheck流程图](/resources/xcheck_process.jpg)

### 运行扫描

点击 `View` > `Tool Windows` > `OpenSCA` 可打开OpenSCA窗口。首先在OpenSCA配置界面中配置服务器参数（参考：插件功能-配置），然后在OpenSCA窗口中点击“运行”（参考：插件功能-运行）