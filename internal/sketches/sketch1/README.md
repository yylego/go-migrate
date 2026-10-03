## 🧾 Module Overview (English)

This module demonstrates loading SQL migration scripts from `scripts/` on disk using the `file` source provided in [`golang-migrate`](https://github.com/golang-migrate/migrate).

### Features

- Uses the `file://` scheme to read migration scripts from disk;
- Iterates through migration versions and reads both `up` and `down` SQL files;
- Supports debugging and checking migration scripts during development.

The `TestFileOpen` function uses `source.Driver` to traverse migration versions and load SQL contents in sequence.

[中文说明](README.zh.md)
