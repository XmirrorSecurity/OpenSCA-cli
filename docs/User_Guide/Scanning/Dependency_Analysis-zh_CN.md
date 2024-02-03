[返回目录](/docs/README-zh-CN.md) / [English](./Dependency_Analysis.md)

# 依赖分析

OpenSCA 通过扫描项目依赖特征文件(动态 + 静态)，生成项目依赖关系。所谓动态解析是指调用包管理器，获取项目依赖关系；静态解析是通过模拟包管理器的行为，获取项目依赖关系。通常情况下，动态解析的结果更加准确，但是需要安装包管理器。静态解析的结果可能不够准确，但是不需要安装包管理器。

依赖分析的目的是帮助用户了解项目的依赖关系，以便更好地管理项目。

支持的语言和包管理器详见 [关于 OpenSCA](/docs/About_OpenSCA-zh_CN.md)

# 使用 OpenSCA-cli 进行依赖分析

## 分析本地项目目录

### 基本命令

 ```shell
 opensca-cli -path {项目路径}
 ```

### 示例

<table>
<tr>
<th align="center">分析 `~/workspace/myproject` 目录</th>
<th align="center">分析 `~/workspace/myproject` 目录并生成报告</th>
</tr>
<tr>
<td>

```shell
opensca-cli -path ~/workspace/myproject
```
</td>
<td>

```shell
opensca-cli -path ~/workspace/myproject -out ~/workspace/myproject/report.html
```

</td>
</tr>
</table>

 ## 分析依赖特征文件

### 基本命令

 ```shell
 opensca-cli -path {依赖特征文件路径}
 ```

### 示例

<table>
<tr>
<th align="center">分析 `~/workspace/myproject/package.json` 文件</th>
<th align="center">分析 `~/workspace/myproject/package.json` 文件并生成报告</th>
</tr>
<tr>
<td>

```shell
opensca-cli -path ~/workspace/myproject/package.json
```

</td>
<td>

```shell
opensca-cli -path ~/workspace/myproject/package.json -out ~/workspace/myproject/report.html
```

</td>
</tr>
</table>

## 分析远程项目

### 基本命令

 ```shell
 opensca-cli -path {项目地址}
 ```

### 示例

<table>
<tr>
<th align="center">分析 ftp 目录</th>
