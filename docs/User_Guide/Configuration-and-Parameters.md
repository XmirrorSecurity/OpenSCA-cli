[Back to Contents](/docs/README.md) | [简体中文](./Configuration-and-Parameters-zh_CN.md)

- [Command-line Parameters](#command-line-parameters)
- [Configuration File](#configuration-file)
- [Ignore Path Configuration](#ignore-path-configuration)

# Command-line Parameters

| Parameter | Description | Example |
| --------- | ----------- | ------- |
| `config` | Set the configuration file path | `-config config.json` |
| `path` | Set the target path. HTTP(S), FTP, and file paths are supported | `-path ./foo` |
| `out` | Set report output paths. File types are detected by suffix | `-out out.json,out.html` |
| `log` | Set the log file path | `-log my_log.txt` |
| `token` | Cloud service token | `-token xxx` |
| `proj` | SaaS project token | `-proj xxx` |
| `version` | Print version information | `-version` |
| `help` | Print help information | `-help` |

# Configuration File

The configuration file uses JSON syntax and supports the following top-level fields:

- `path`: `String` target path. HTTP(S), FTP, and file paths are supported.
- `out`: `String` report output paths. Supported suffixes include html/json/xml/csv/sqlite/cdx/spdx/swid/dsdx.
- `optional`: `Object` optional scanning settings.
  - `ui`: `Boolean` enable the interactive UI. Default: `false`.
  - `dedup`: `Boolean` deduplicate identical components and merge paths. Default: `false`.
  - `dir`: `Boolean` scan directories only and skip archives. Default: `false`.
  - `vuln`: `Boolean` keep only vulnerable components. Default: `false`.
  - `progress`: `Boolean` show the progress bar. Default: `true`.
  - `dev`: `Boolean` keep development dependencies. Default: `true`.
  - `tls`: `Boolean` enable TLS certificate verification. Default: `false`.
  - `proxy`: `String` HTTP proxy address. Default: empty.
  - `ignore`: `Array<String>` path rules ignored during scanning. Default: empty. OpenSCA only reads these rules from the current configuration file and does not automatically load the project's `.gitignore`. The syntax is compatible with common `.gitignore` rules, including directory matches, wildcards, and `!` negation.
- `repo`: `Object` component repository settings for Maven, npm, and Composer.
- `origin`: `Object` vulnerability database settings.

# Ignore Path Configuration

Use `optional.ignore` to skip test dependencies, temporary directories, or specific archives:

```json
{
  "optional": {
    "ignore": [
      "JarCollection/",
      "*.jar",
      "!libs/keep.jar"
    ]
  }
}
```

The example above skips `JarCollection/` and all `.jar` files, but keeps `libs/keep.jar`. Ignore rules only affect OpenSCA scanning and do not modify project files.
