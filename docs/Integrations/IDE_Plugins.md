[Go Back](/docs/README.md) | [简体中文](./IDE_Plugins-zh_CN.md)

- [Visual Studio Code](#visual-studio-code)
  - [Install Plugin](#install-plugin)
  - [Using the Plugin](#using-the-plugin)
    - [Plugin Features](#plugin-features)
    - [Plugin Execution Process](#plugin-execution-process)
    - [Running a Scan](#running-a-scan)
- [JetBrains IDEs](#jetbrains-ides)
  - [Installing the Plugin](#installing-the-plugin)
  - [Using the Plugin](#using-the-plugin-1)
    - [Plugin Features](#plugin-features-1)
    - [Plugin Execution Process](#plugin-execution-process-1)
    - [Running a Scan](#running-a-scan-1)

# Visual Studio Code

## Install Plugin

- **Option 1**：Install from [VS Marketplace](https://marketplace.visualstudio.com/items?itemName=xmirror.opensca)

    In VS Code, open Extensions in the left sidebar -> enter "OpenSCA Xcheck" in the extension search box, click "Install"

    <img src="https://opensca.xmirror.cn/docs/img/vscode_01.jpg" alt="xcheck_market" />

- **Option 2**：Download from [OpenSCA Official Web Site](https://opensca.xmirror.cn/pages/plug-in)

  - Download "OpenSCA-Xcheck.vsix" from the OpenSCA official website;
  - Open VS Code, open Extensions in the left sidebar -> more actions in the top bar of the extension -> "Install from VSIX" -> find and install "OpenSCA-Xcheck.vsix";

- **Option 3**：Build from source code

  - Requirements:
    - node v18 and above

  - Clone the repository from [github](https://github.com/XmirrorSecurity/OpenSCA-VSCode-plugin/) or [gitee](https://gitee.com/XmirrorSecurity/OpenSCA-VSCode-plugin)

  * Install vsce

    ```
    npm install --global @vscode/vsce
    ```

  * Package

    ```
    vsce package
    ```

## Using the Plugin

### Plugin Features

- **Start Scan**: Click the "Run" button in the action bar to start scanning for vulnerabilities in the components of the current project.
- **Stop Scan**: Click the "Stop" button in the action bar to stop the ongoing scan for vulnerabilities in the current project.
- **Clear Scan Results**: Click the "Clean" button in the action bar to clear the scan results of the current project.
- **Connection Configuration**: Click the "Test" button in the action bar to configure the platform URL and Token information. Click the "Test Connection" button to verify if the connection configuration is correct. Once the connection is successful, you can start scanning.
- **Settings**: Click the "Setting" button in the action bar to view the settings related to OpenSCA Xcheck.
- **Instructions**: Click the "Instructions" button in the action bar to view the user manual for OpenSCA Xcheck.
- **See More**: Click the "See more" button in the action bar to visit [opensca.xmirror.cn](https://opensca.xmirror.cn) for more information about OpenSCA Xcheck.

![xcheck_function](https://opensca.xmirror.cn/docs/img/vscode_02.jpg)

### Plugin Execution Process

![xcheck_flow](https://opensca.xmirror.cn/docs/assets/img/xcheck_process.7083b869.jpg)

### Running a Scan

Click on OpenSCA Xcheck to open the OpenSCA window. First, configure the server parameters in the configuration interface (refer to: Plugin Features - Settings), then click “Run” in the OpenSCA window (refer to: Plugin Features - Start Scan).

# JetBrains IDEs

## Installing the Plugin

- **Method 1**: Install from the [JetBrains Plugin Marketplace](https://plugins.jetbrains.com/plugin/18246-opensca-xcheck) (Recommended)

    For example, in IntelliJ IDEA: go to `File | Settings | Plugins | Marketplace`, search for "OpenSCA Xcheck" in the search box, and click "Install".
    
    ![xcheck_market](https://opensca.xmirror.cn/docs/img/xcheck_marketplace.jpg)

- **Method 2**: Download the plugin from the [OpenSCA Platform](https://opensca.xmirror.cn/pages/plug-in) and install it manually

    For example, in IntelliJ IDEA: drag the downloaded plugin package into the IDE.

- **Method 3**: [Download the source code](https://github.com/XmirrorSecurity/OpenSCA-intellij-plugin) and compile it yourself

    Open the downloaded source code in IntelliJ IDEA. Configure the runtime environment: `JDK11`. After Gradle imports dependencies and plugins, execute the `buildPlugin` task of the `intellij` plugin in Gradle. The built package will be located in the `build/distributions` directory of the project. Drag this package into the IDE to install it.

## Using the Plugin

### Plugin Features

- **Configuration**: Click `File | Settings | Other Settings | OpenSCA Setting` or click the `Setting` button in the OpenSCA window to configure the server URL and Token in the configuration interface.
- **Test Connection**: After configuring the server URL and Token in the OpenSCA configuration interface, click the `Test Connection` button to verify if the URL and Token are valid.
- **Run**: Click the `Run` button in the OpenSCA window to perform a code assessment on the current project.
- **Stop**: If a code assessment is ongoing for the current project, the `Stop` button will be enabled. Click the `Stop` button to end the current assessment task.
- **Clear**: If the Xcheck sub-window in the OpenSCA window already has assessment results, click the `Clean` button to clear all results in the Xcheck sub-window.

![xcheck_function](https://opensca.xmirror.cn/docs/img/xcheck_function.jpg)

### Plugin Execution Process

![xcheck_flow](https://opensca.xmirror.cn/docs/img/xcheck_process.jpg)

### Running a Scan

Click `View > Tool Windows > OpenSCA` to open the OpenSCA window. First, configure the server parameters in the OpenSCA configuration interface (refer to: Plugin Features - Configuration), then click the "Run" button in the OpenSCA window (refer to: Plugin Features - Run).