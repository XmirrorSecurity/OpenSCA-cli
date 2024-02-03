[Go Back](/docs/README.md) | [简体中文](./CICD-zh_CN.md)

- [Jenkins](#jenkins)
  - [Freestyle Project](#freestyle-project)
  - [Pipeline Project](#pipeline-project)
  - [(Optional) Post-build Actions](#optional-post-build-actions)
- [GitLab CI](#gitlab-ci)
- [GitHub Actions](#github-actions)

# Jenkins

To integrate OpenSCA into Jenkins, it is a must to install [OpenSCA-cli](https://github.com/XmirrorSecurity/OpenSCA-cli) in Jenkins build agent. OpenSCA supports major OS including Windows, Linux and MacOS, as well as docker image.

## Freestyle Project

Add `Execute shell` or `Execute Windows batch command` to the building process to run OpenSCA-cli.

![jenkins_freestyle](/resources/jenkins-freestyle.png)

```bash
# install opensca-cli
curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh
# export opensca-cli to PATH
export PATH=/var/jenkins_home/.config/opensca-cli:$PATH
# run opensca scan and generate reports(replace {put_your_token_here} with your token)
opensca-cli -path $WORKSPACE -token {put_your_token_here} -out $WORKSPACE/results/result.html,$WORKSPACE/results/result.dsdx.json
```

## Pipeline Project

Add `sh` or `bat` to the pipeline script to run OpenSCA-cli.

```groovy
pipeline {
    agent any

    stages {

        stage('Build') {
            steps {
                // Get some code from a GitHub repository
                // build it, test it, and archive the binaries.
            }
        }

        stage('Security Scan') {
            steps {
                // install opensca-cli
                sh "curl -sSL https://raw.githubusercontent.com/XmirrorSecurity/OpenSCA-cli/master/scripts/install.sh | sh"
                // run opensca scan and generate reports(replace {put_your_token_here} with your token)
                sh "/var/jenkins_home/.config/opensca-cli/opensca-cli -path $WORKSPACE -token {put_your_token_here} -out $WORKSPACE/results/result.html,$WORKSPACE/results/result.dsdx.json"
            }
        }
    }

    post {
        always {
            // do something post build
        }
    }
}

```

## (Optional) Post-build Actions

Show the HTML report output by OpenSCA via `Publish HTML reports` plugin.

> Enabling JavaScript is a prerequisite to show the report properly. That needs to [adjust the security policy of Jenkins](https://www.jenkins.io/doc/book/security/configuring-content-security-policy//). Please be cautious given that such adjustment may weaken the security of Jenkins.

<details>
<summary>Change Jenkins CSP</summary>

Execute the following script in `Manage Jenkins` -> `Script Console` ：

```groovy
System.setProperty("hudson.model.DirectoryBrowserSupport.CSP", "sandbox allow-scripts; default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval';")
```

Restart Jenkins after execution.

</details>

Making sure the `Publish HTML reports` plugin has been installed, and then add `Publish HTML reports` to the Jenkins project's `Post-build Actions` ：

![jenkins_postbuild](/resources/jenkins-postbuild.png)  

When the build is succeeded, the HTML output will be available in the Dashboard of Jenkins Job

![html_report](/resources/jenkins-view-html-report.gif)

<details>
<summary>Pipeline Script Example</summary>

```groovy
post {
    always {
        // do something post build
        publishHTML(
            [
                allowMissing: false,
                alwaysLinkToLastBuild: true,
                keepAll: true,
                reportDir: 'results',
                reportFiles: 'result.html',
                reportName: 'OpenSCA Report',
                reportTitles: 'OpenSCA Report',
                useWrapperFileDirectly: true
            ]
        )
    }
}
```

</details>


# GitLab CI
Install [OpenSCA-cli](https://github.com/XmirrorSecurity/OpenSCA-cli) in GitLab Runner to integrate OpenSCA. OpenSCA supports major OS including Windows, Linux and MacOS, as well as docker image.

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
<summary> Complete Example </summary>

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

Integrate OpenSCA into GitHub Actions with the help of [OpenSCA Scan Action](https://github.com/marketplace/actions/opensca-scan-action).

```yaml
name: OpenSCA Scan

on: 
  push:
    branches: 
      - master
  pull_request:
    branches: 
      - master

jobs:
  opensca_scan:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: OpenSCA Scan
        uses: XmirrorSecurity/opensca-scan-action@v1.0.0
        with:
          token: ${{ secrets.OPENSCA_TOKEN }}
```

For more information, please check [OpenSCA Scan Action](https://github.com/XmirrorSecurity/opensca-scan-action)
