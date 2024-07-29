[返回目录](/docs/README-zh-CN.md) / [English](./Reports.md)

- [生成报告](#生成报告)
- [使用 OpenSCA-cli 生成报告](#使用-opensca-cli-生成报告)
  - [基本命令](#基本命令)
  - [示例](#示例)
- [使用 json2excel 脚本生成 Excel 格式报告](#使用-json2excel-脚本生成-excel-格式报告)
  - [需求](#需求)
  - [安装脚本](#安装脚本)
  - [使用方法](#使用方法)

# 生成报告

OpenSCA 支持生成多种格式的报告, 包括 JSON(`.json`), XML(`.xml`), HTML(`.html`), SQLite(`.sqlite`), CSV(`.csv`), SARIF(`.sarif`)

> CSV 格式报告仅包含依赖关系, 不包含漏洞信息. 若需要包含漏洞信息的 Excel 格式报告, 可通过 [json2excel](https://github.com/XmirrorSecurity/OpenSCA-cli/blob/master/scripts/json2excel.py) 脚本, 将 JSON 格式报告转换为 Excel 格式报告.

此外, OpenSCA 提供 SaaS 服务, 同步扫描结果后, 可以在 [OpenSCA SaaS Console](https://opensca.xmirror.cn/console) 查看和下载报告.

# 使用 OpenSCA-cli 生成报告

## 基本命令

OpenSCA-cli 使用 `-out` 参数指定报告输出路径, 使用后缀名指定报告格式.

```shell
opensca-cli -path {项目路径} -out {报告路径}.{报告格式}
```

`-out` 参数支持指定多个报告路径, 使用半角逗号(`,`)分隔.

```shell
opensca-cli -path {项目路径} -out {报告路径1}.{报告格式1},{报告路径2}.{报告格式2}
```

## 示例

<table>
<tr>
<th align="center">生成 JSON 格式报告</th>
<th align="center">生成 XML 格式报告</th>
<th align="center">生成 HTML 格式报告</th>
<th align="center">生成 SQLite 格式报告</th>
<th align="center">生成 CSV 格式报告</th>
<th align="center">生成 SARIF 格式报告</th>
</tr>
<tr>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/report.json
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/report.xml
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/report.html
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/report.sqlite
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/report.csv
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/report.sarif
```

</td>
</tr>
</table>

# 使用 json2excel 脚本生成 Excel 格式报告

## 需求

- Python 3.6+
- pandas

## 安装脚本

```shell
# 下载脚本
wget https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/json2excel.py

# 安装依赖
pip install pandas
```

## 使用方法

**修改 json2excel.py**

修改 json2excel.py 文件中的以下内容: 

- 将 "result.json" 修改为你的 JSON 格式报告路径
- 将 "result.xlsx" 修改为你的 Excel 格式报告路径

```python
# ...
if __name__ == "__main__":
    json2excel("result.json", "result.xlsx")
```

**运行脚本**

```shell
python json2excel.py
```
