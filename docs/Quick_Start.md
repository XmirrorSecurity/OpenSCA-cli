[Go Back](/docs/README.md) | [中文](./Quick_Start-zh_CN.md)

# Quick Start

## Video

<video width="320" height="240" controls="controls" poster="https://opensca.xmirror.cn/docs/assets/img/poster_cli.d9973be2.png" style="width: 100%; max-height: 500px; height: auto;" jm_neat="328132609"><source src="https://opensca.xmirror.cn/docs/assets/media/cli.1bed8c1c.mp4" type="video/mp4">
您的浏览器不支持 video 标签。
</video>

## Traditional Method

### Download & Deploy

- Method No.1：one-click installation
  - For Mac/Linux Users
    ```shell
    curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh

    # Try this when internet connection fails
    curl -sSL https://gitee.com/XmirrorSecurity/OpenSCA-cli/raw/master/scripts/install.sh | sh -s -- gitee
    ```
- Method No.2：via `Homebrew` (Mac/Linux)
  ```shell
  brew install opensca-cli
  ```
- Method No.3：Download executable program and decompress from [GitHub](https://github.com/XmirrorSecurity/OpenSCA-cli/releases/latest) or [Gitee](https://gitee.com/XmirrorSecurity/OpenSCA-cli/releases/latest) 

### Use OpenSCA 

**Report dependencies of the given directory**

```shell
opensca-cli -path ${project_path}
```

**Report dependencies and get info of vuls & licenses from cloud knowledge base**

> Register first [register](https://opensca.xmirror.cn/register) and get token

```shell
opensca-cli -path ${project_path} -token ${token}
```

## Docker

**Report dependencies of the given directory**

```shell
docker run -ti --rm -v ${project_path}:/src opensca/opensca-cli:latest
```

**Report dependencies and get info of vuls & licenses from cloud knowledge base**

> Register first [register](https://opensca.xmirror.cn/register) and get token

```shell
docker run -ti --rm -v ${project_path}:/src opensca/opensca-cli:latest -token ${token}
```