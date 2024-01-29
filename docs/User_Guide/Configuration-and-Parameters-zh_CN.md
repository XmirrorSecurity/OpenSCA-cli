[返回目录](/docs/README-zh-CN.md) | [English](./Configuration-and-Parameters.md)

# 命令行参数

| 参数      | 描述                                         | 使用示例                 |
| --------- | -------------------------------------------- | ------------------------ |
| `config`  | 指定配置文件路径                             | `-config config.json`    |
| `path`    | 指定检测项目路径, 支持 http(s)/ftp/file 协议 | `-path ./foo`            |
| `out`     | 根据后缀生成报告                             | `-out out.json,out.html` |
| `log`     | 指定日志文件路径                             | `-log my_log.txt`        |
| `token`   | 云端服务`token`                              | `-token xxx`             |
| `proj`    | saas项目`token`                              | `-proj xxx`              |
| `version` | 显示版本信息                                 | `-version`               |
| `help`    | 显示帮助信息                                 | `-help`                  |

# 配置文件说明

配置文件使用 `json` 格式，支持以下字段: 
> 默认会从目标检测路径中查找配置文件, 否则使用[默认配置文件](/config.json)。 可通过 `-config` 参数指定配置文件路径。

- `path`: `String` 检测目标路径, 支持 http(s)/ftp/file 协议
- `out`: `String` 报告输出路径, 通过后缀名识别文件类型, 支持 html/json/xml/csv/sqlite/cdx/spdx/swid/dsdx
- `optional`: `Object` 可选配置项
  - `ui`: `Boolean` 是否启用交互式界面, 默认为 `false`
  - `dedup`: `Boolean` 是否启用组件去重(相同组件仅保留一条记录，组件路径合并), 默认为 `false`
  - `dir`: `Boolean` 是否仅检测目录(跳过压缩包), 默认为 `false`
  - `vuln`: `Boolean` 是否仅保留漏洞组件, 默认为 `false`
  - `progress`: `Boolean` 是否显示进度条, 默认为 `true`
  - `dev`: `Boolean` 是否保留开发组件, 默认为 `true`
  - `tls`: `Boolean` 开启 TLS 证书验证, 默认为 `false`
  - `proxy`: `String` 代理地址, 默认为空
- `repo`: `Object` 组件仓库配置
  - `maven`: `Array` maven 镜像/私服仓库配置
    - `url`: `String` 仓库地址
    - `user`: `String` 用户名
    - `pass`: `String` 密码
  - `npm`: `Array` npm 镜像/私服仓库配置
    - `url`: `String` 仓库地址
    - `user`: `String` 用户名
    - `pass`: `String` 密码  
  - `composer`: `Array` composer 镜像/私服仓库配置
    - `url`: `String` 仓库地址
    - `user`: `String` 用户名
    - `pass`: `String` 密码
- `origin`: `Object` 漏洞数据源配置
  - `url`: `String` 漏洞数据源地址
  - `token`: `String` 云端漏洞数据库个人访问令牌
  - `proj`: `String` 项目访问令牌, 若置空则同步结果至"快速检测", 若无此字段(注释或删除)则不将结果同步至 OpenSCA SaaS
  - `json`: `String` JSON 格式漏洞数据库路径
  - `mysql`: `Object` MySQL 数据库漏洞数据源配置
    - `dsn`: `String` 数据库连接字符串
    - `table`: `String` 数据表名
  - `sqlite`: `Object` SQLite 数据库漏洞数据源配置
    - `dsn`: `String` 数据库连接字符串
    - `table`: `String` 数据表名
