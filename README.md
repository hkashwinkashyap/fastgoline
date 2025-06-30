# fastgoline

**FastGoLine** is a high-performance data pipeline processing library built in Go.

---

## ✨ Features

* **Generic Type Support** – Works with any data type using Go generics.

* **Concurrent Stage Execution** – Each pipeline stage runs in its own goroutine.

* **Stage Reusability** – Reuse transformation logic across pipelines.

* **Pipeline Composition** – Chain multiple stages to transform input to output.

* **Deadlock-Safe** – Designed to prevent blocking and goroutine leaks.

* **Multi-Pipeline Execution** – Run multiple pipelines in parallel with shared or unique stages.

* **UUID-based Identification** – Every pipeline and stage have their unique ID for tracing.


## 📦 Installation

```bash
go get github.com/hkashwinkashyap/fastgoline
```

## 🚀 Usage

Refer the `main.go` file to define and run multiple pipelines concurrently.

NOTE: The environment variables can be set on the machine and when `fgl_config.InitialiseConfig()` is called, the values will be pulled in.


## 📄 License

This project is licensed under the [MIT License](./LICENSE).
