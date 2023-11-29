# 代码规范

## 命名

| 类型   | 格式                                           | 示例                     |
| ------ | ---------------------------------------------- | ------------------------ |
| 常量   | 全大写                                         | `MAX_COUNT`              |
| 变量名 | 驼峰，可适当缩写                               | `taskId, TaskId`         |
| 函数   | 驼峰                                           | `checkToken, CheckToken` |
| 结构   | 驼峰                                           | `fileInfo, FileInfo`     |
| 接口   | 驼峰，一般`er`结尾                             | `Analyzer`               |
| 文件名 | 全小写，下划线分割                             | `db_test.go`             |
| 包名   | 全小写，尽可能使用简短的单词，避免下划线或驼峰 | `config`                 |

## 注释

- 函数 使用单行注释(`//`)，函数名开头。

  ```go
  // RunTask is run a task
  // taskId is task id
  func RunTask(taskId int64)
  ```

- 代码 无法直观了解逻辑的代码需要添加注释。