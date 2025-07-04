# fastgoline

**FastGoLine** is a high-performance data pipeline processing library built in Go.

## ✨ Features

* **Generic Type Support** – Works seamlessly with any data type using Go generics.
* **Concurrent Stage Execution** – Each pipeline stage runs independently in its own goroutine for maximum throughput.
* **Stage Reusability** – Easily reuse transformation logic across different pipelines.
* **Pipeline Composition** – Chain multiple stages to flexibly transform input to output.
* **Deadlock-Safe** – Engineered to avoid blocking and goroutine leaks for reliable long-running use.
* **Multi-Pipeline Execution** – Run multiple pipelines in parallel, sharing or isolating stages as needed.
* **UUID-based Identification** – Unique IDs assigned to every pipeline and stage for effortless tracing and debugging.
* **Stress Tested** – Proven stable and performant under heavy load scenarios.

## 📦 Installation

```bash
go get github.com/hkashwinkashyap/fastgoline
```

## 🚀 Usage

Check out the `main.go` example to see how to define and run multiple pipelines concurrently.

**Note:** Environment variables can be set on your system and will be automatically loaded when calling `fgl_config.InitialiseConfig()`.

## 📄 License

This project is licensed under the [MIT License](./LICENSE).