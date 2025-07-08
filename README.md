# fastgoline

**FastGoLine** is a high-performance data pipeline processing library built in Go.

## :sparkles: Features

* :envelope: **Generic Type Support** – Works seamlessly with any data type using Go generics.
* :gear: **Concurrent Stage Execution** – Each pipeline stage runs independently in its own goroutine for maximum throughput.
* :recycle: **Stage Reusability** – Easily reuse transformation logic across different pipelines.
* :factory: **Pipeline Composition** – Chain multiple stages to flexibly transform input to output.
* :lock: **Deadlock-Safe** – Engineered to avoid blocking and goroutine leaks for reliable long-running use.
* :chains: **Multi-Pipeline Execution** – Run multiple pipelines in parallel, sharing or isolating stages as needed.
* :ticket: **UUID-based Identification** – Unique IDs assigned to every pipeline and stage for effortless tracing and debugging.
* :memo: **Stress Tested** – Proven stable and performant under heavy load scenarios.

## :package: Installation

```bash
go get github.com/hkashwinkashyap/fastgoline
```

## :rocket: Usage

Check out the `main.go` example to see how to define and run multiple pipelines concurrently.

**Note:** Environment variables can be set on your system and will be automatically loaded when calling `fgl_config.InitialiseConfig()`.

**Note:** `fgl_stage.NewStageFunction()` can be defined as required but **make sure to cast the ouput value** to match the `Generic Type` of the Stage and Pipeline.

## :page_facing_up: License

This project is licensed under the [MIT License](./LICENSE).