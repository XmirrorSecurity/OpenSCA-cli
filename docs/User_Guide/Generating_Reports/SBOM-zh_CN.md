[返回目录](/docs/README-zh-CN.md) / [English](./SBOM.md)

- [SBOM 清单](#sbom-清单)
- [生成 SBOM 清单](#生成-sbom-清单)
  - [基本命令](#基本命令)
  - [示例](#示例)
- [转换 SBOM 清单](#转换-sbom-清单)
  - [基本命令](#基本命令-1)
  - [示例](#示例-1)


# SBOM 清单

OpenSCA 支持生成多种格式的 SBOM 清单, 包括 [DSDX](https://opensca.xmirror.cn/resources/particulars?id=133), [SPDX](https://spdx.dev/), [CycloneDX](https://cyclonedx.org/) 和 [SWID](https://csrc.nist.gov/projects/Software-Identification-SWID).

除了生成 SBOM, OpenSCA 还支持将 SBOM 作为输入, 生成漏洞报告或转换为其他格式的 SBOM 清单.

支持的 SBOM 格式:

| SBOM 清单   | 文件后缀                         |
| ----------- | -------------------------------- |
| `DSDX`      | `.dsdx` `.dsdx.json` `.dsdx.xml` |
| `SPDX`      | `.spdx` `.spdx.json` `.spdx.xml` |
| `CycloneDX` | `.cdx.json` `.cdx.xml`           |
| `SWID`      | `.swid.json` `.swid.xml`         |

# 生成 SBOM 清单

## 基本命令

OpenSCA-cli 使用 `-out` 参数指定 SBOM 清单输出路径, 使用后缀名指定 SBOM 清单格式.

```shell
opensca-cli -path {项目路径} -out {SBOM路径}.{SBOM格式}
```

`-out` 参数支持指定多个 SBOM 路径, 使用半角逗号(`,`)分隔.

```shell
opensca-cli -path {项目路径} -out {SBOM路径1}.{SBOM格式1},{SBOM路径2}.{SBOM格式2}
```

也可以同时生成 SBOM 清单和漏洞报告.

```shell
opensca-cli -path {项目路径} -out {SBOM路径}.{SBOM格式},{报告路径}.{报告格式}
```

## 示例

<table>
<tr>
<th align="center">生成 JSON 格式 DSDX 清单</th>
<th align="center">生成 JSON 格式 SPDX 清单</th>
<th align="center">生成 JSON 格式 CycloneDX 清单</th>
<th align="center">生成 JSON 格式 SWID 清单</th>
</tr>
<tr>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/sbom.dsdx.json
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/sbom.spdx.json
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/sbom.cdx.json
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject -out ~/workscapce/myproject/sbom.swid.json
```

</td>
</tr>
</table>

# 转换 SBOM 清单

## 基本命令

```shell
opensca-cli -path {SBOM路径}.{SBOM格式} -out {SBOM路径}.{SBOM格式}
```

> 注意: SPDX 和 SWID 格式的 SBOM 清单不包含组件语言信息或包管理器坐标, 在转换时可能会丢失部分信息.

## 示例

<table>
<tr>
<th align="center">DSDX 转 SPDX</th>
<th align="center">DSDX 转 CycloneDX</th>
<th align="center">DSDX 转 SWID</th>
</tr>
<tr>
<td>

```shell
opensca-cli -path ~/workscapce/myproject/sbom.dsdx.json -out ~/workscapce/myproject/sbom.spdx.json
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject/sbom.dsdx.json -out ~/workscapce/myproject/sbom.cdx.json
```

</td>
<td>

```shell
opensca-cli -path ~/workscapce/myproject/sbom.dsdx.json -out ~/workscapce/myproject/sbom.swid.json
```

</td>
</tr>
</table>
