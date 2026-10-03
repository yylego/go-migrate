## 🧾 Module Overview (English)

This module demonstrates embedding SQL migration scripts into a Go program using `embed.FS`, then reading and executing them using the `iofs` source from `golang-migrate`.

Benefits of this approach:

- Migration scripts are versioned with the code;
- Simplified deployment without separate `.sql` files;
- Supports automated tests and CI/CD.

Two embedding strategies are shown:

- Embedding the complete `scripts/` tree;
- Embedding selected `.sql` files.

Choose the complete migration sequence when deploying; choose a subset when testing.

[中文说明](README.zh.md)
