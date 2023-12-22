---
title: CI/CD
author: Cyber Chen
data: 2023-12-20T17:17:00+08:00
---

[返回目录](/docs/README-zh-CN.md) | [English](./CICD.md)

- [Jenkins](#jenkins)
  - [Freestyle Project](#freestyle-project)
  - [Pipeline Project](#pipeline-project)
  - [(Optional) Post-build Actions](#optional-post-build-actions)
- [GitLab CI](#gitlab-ci)
- [GitHub Actions](#github-actions)

# Jenkins

在 Jenkins 中集成 OpenSCA，需要在 Jenkins 构建机器中安装 [OpenSCA-cli](https://github.com/XmirrorSecurity/OpenSCA-cli)。 OpenSCA-cli 支持主流的操作系统，包括 Windows、Linux、MacOS，亦可通过 Docker 镜像运行。

## Freestyle Project

对于自由风格的项目，可以通过在构建步骤中添加 `Execute shell` 或 `Execute Windows batch command` 来执行 OpenSCA-cli。

![jenkins_freestyle](/resources/jenkins-freestyle.png)

```bash
# install opensca-cli
curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh
# export opensca-cli to PATH
export PATH=/var/jenkins_home/.config/opensca-cli:$PATH
# run opensca scan and generate reports（replace {put_your_token_here} with your token）
opensca-cli -path $WORKSPACE -token {put_your_token_here} -out $WORKSPACE/results/result.html,$WORKSPACE/results/result.dsdx.json
```

## Pipeline Project

对于流水线项目，可以通过在流水线脚本中添加 `sh` 或 `bat` 来执行 OpenSCA-cli。

```groovy
// TODO
```

## (Optional) Post-build Actions

在 Jenkins 中，可以通过 `Publish HTML reports` 插件来展示 OpenSCA-cli 生成的 HTML 报告。

> 请注意，OpenSCA 生成的 HTML 报告需启用 JavaScript 才能正常显示。这需要修改 Jenkins 的安全策略，具体操作请参考 [Jenkins 官方文档](https://www.jenkins.io/doc/book/security/configuring-content-security-policy//)。这可能会导致 Jenkins 的安全性降低，因此请谨慎操作。

<details>
<summary>修改 Jenkins CSP</summary>

在 Jenkins 的 `Manage Jenkins` -> `Script Console` 中执行以下脚本：

```groovy
System.setProperty("hudson.model.DirectoryBrowserSupport.CSP", "sandbox allow-scripts; default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval';")
```

执行完成后，需重启 Jenkins 服务。

</details>

确保您已经安装了 `Publish HTML reports` 插件，然后在 Jenkins 项目的 `Post-build Actions` 中添加 `Publish HTML reports`：

![jenkins_postbuild](/resources/jenkins-postbuild.png)  

成功构建后，在 Jenkins Job 的 Dashboard 中，即可看到 OpenSCA-cli 生成的 HTML 报告

![html_report](/resources/jenkins-view-html-report.gif)


# GitLab CI

在 GitLab CI 中集成 OpenSCA，需要在 GitLab Runner 中安装 [OpenSCA-cli](https://github.com/XmirrorSecurity/OpenSCA-cli)。 OpenSCA-cli 支持主流的操作系统，包括 Windows、Linux、MacOS，亦可通过 Docker 镜像运行。

```yaml
security-test-job:
    stage: test
    script:
        - echo "do opensca scan..."
        - curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh
        - /root/.config/opensca-cli/opensca-cli -path $CI_PROJECT_DIR -token {put_your_token_here} -out $CI_PROJECT_DIR/results/result.html,$CI_PROJECT_DIR/results/result.dsdx.json
    artifacts:
      paths:
        - results/
      untracked: false
      when: on_success
      expire_in: 30 days
```

<details>
<summary>完整示例</summary>

```yaml
stages:
  - build
  - test
  - deploy

build-job:
  stage: build
  script:
    - echo "Compiling the code..."
    - echo "Compile complete."

unit-test-job:
  stage: test
  script:
    - echo "do unit test..."
    - sleep 10
    - echo "Code coverage is 90%"

lint-test-job:
  stage: test
  script:
    - echo "do lint test..."
    - sleep 10
    - echo "No lint issues found."

security-test-job:
    stage: test
    script:
        - echo "do opensca scan..."
        - curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh
        - /root/.config/opensca-cli/opensca-cli -path $CI_PROJECT_DIR -token {put_your_token_here} -out $CI_PROJECT_DIR/results/result.html,$CI_PROJECT_DIR/results/result.dsdx.json
    artifacts:
      paths:
        - results/
      untracked: false
      when: on_success
      expire_in: 30 days

deploy-job:
  stage: deploy
  environment: production
  script:
    - echo "Deploying application..."
    - echo "Application successfully deployed."
```

</details>

# GitHub Actions
