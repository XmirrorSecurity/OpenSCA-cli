# Code Standard

## Naming  

| TYPE         | FORMAT                                                       | EXAMPLE                  |
| ------------ | ------------------------------------------------------------ | ------------------------ |
| Constant     | all in upper case                                            | `MAX_COUNT`              |
| Variables    | camel-case; proper abbreviation is acceptable                | `taskId, TaskId`         |
| Function     | camel-case                                                   | `checkToken, CheckToken` |
| Structure    | camel-case                                                   | `fileInfo, FileInfo`     |
| API          | camel-case; end with `er`                                    | `Analyzer`               |
| Doc name     | all in lower case; use the underscore to show separation     | `db_test.go`             |
| Package name | all in lower case; be brief and avoid using the underscore or camel-case if possible | `config`                 |

## Comments

- Function: Please use single-line comments(`//`), starting with the name of the function.

  ```go
  // RunTask is run a task
  // taskId is task id
  func RunTask(taskId int64)
  ```

- Code: We recommend adding comments to your code so that we can grasp the underlying logic more effectively.