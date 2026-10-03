## 🧾 模块简介

使用 `embed.FS` 加载迁移脚本，通过 `source.DefaultParse` 解析文件名，利用 `golang-migrate` 的 `source.NewMigrations()` 构建迁移列表。

主要内容：

- 通过 `fs.ReadDir` 读取嵌入的迁移脚本；
- 解析文件名，提取版本、方向等迁移元数据；
- 将解析结果添加到 `Migrations` 结构；
- 验证 `.First()`、`.Up()` 和 `.Down()` 方法，模拟迁移逻辑。

适用场景：

- 编写自定义迁移逻辑；
- 不依赖额外服务，测试文件名解析；
- 理解 `golang-migrate` 内部的迁移注册流程。

[English](README.md)
