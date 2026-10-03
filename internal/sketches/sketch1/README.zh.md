## 🧾 模块简介

演示使用 [`golang-migrate`](https://github.com/golang-migrate/migrate) 的 `file` 源，从本地 `scripts/` 读取 SQL 迁移脚本。

### 核心特点

- 使用 `file://` 路径读取本地迁移脚本；
- 遍历迁移版本，读取对应的 `up` 和 `down` SQL 文件；
- 支持开发时调试和检查迁移脚本。

`TestFileOpen` 通过 `source.Driver` 遍历迁移版本，顺序读取 SQL 内容。

[English](README.md)
