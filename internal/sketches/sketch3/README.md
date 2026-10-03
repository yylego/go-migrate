## 🧾 Module Overview (English)

This module uses `embed.FS` to load migration scripts, parses them with `source.DefaultParse`, and builds a migration list with `source.NewMigrations()` from `golang-migrate`.

Features:

- Reads embedded migration scripts with `fs.ReadDir`;
- Parses file names to extract migration metadata (version, direction);
- Appends parsed migrations into a `Migrations` structure;
- Exercises `.First()`, `.Up()`, and `.Down()` methods to simulate migration logic.

Can help you:

- Writing custom migration logic;
- Testing file name parsing without extra services;
- Understanding how `golang-migrate` handles internal migration registration.

[中文说明](README.zh.md)
