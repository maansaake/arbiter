# Arbiter

[![Main branch protection](https://github.com/maansaake/arbiter/actions/workflows/main.yaml/badge.svg)](https://github.com/maansaake/arbiter/actions/workflows/main.yaml)
[![Code scanning](https://github.com/maansaake/arbiter/actions/workflows/code-scanning.yaml/badge.svg)](https://github.com/maansaake/arbiter/actions/workflows/code-scanning.yaml)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/maansaake/arbiter)](https://goreportcard.com/report/github.com/maansaake/arbiter)
![tag](https://img.shields.io/github/v/tag/maansaake/arbiter?label=latest%20version)

Arbiter is a load testing framework for Go. It makes no assumptions about the system under test (SUT) — any protocol or technology works. You describe your workload by implementing a module, then hand it to Arbiter and let it drive traffic.

## Writing a module

A module is a Go struct that implements the `module.Module` interface:

```go
type Module interface {
    Name() string       // unique kebab-case name, e.g. "my-api"
    Desc() string       // human-readable description shown in CLI help
    Args() Args         // module-level configuration flags
    Ops()  Ops          // list of operations Arbiter will call
    Run()  error        // called once before traffic starts
    Stop() error        // called once after traffic stops
}
```

Pass one or more modules to `arbiter.Run` and Arbiter takes care of the rest:

```go
func main() {
    err := arbiter.Run(module.Modules{mymod.New()}, nil)
    // ...
}
```

### Args

Args are typed configuration values that become CLI flags automatically. Supported types are `int`, `uint`, `float64`, `string`, and `bool`.

```go
s.args = module.Args{
    &module.Arg[string]{
        Name:     "host",
        Desc:     "Target host.",
        Required: true,
        Value:    &s.host,
    },
    &module.Arg[int]{
        Name: "timeout-ms",
        Desc: "Request timeout in milliseconds.",
        Handler: func(v int) {
            s.timeout = time.Duration(v) * time.Millisecond
        },
    },
}
```

`Value` holds the parsed result. `Handler` is called after parsing and can be used for additional conversion. Both can be used together. `Required: true` causes Arbiter to fail at startup if the flag is not provided.

### Ops

Ops are the individual operations Arbiter will execute. Each op has a `Do` function and a `Rate` (calls per minute). Arbiter manages scheduling and concurrency.

```go
s.ops = module.Ops{
    &module.Op{
        Name: "get-user",
        Desc: "Fetches a user by ID.",
        Rate: 120, // 120 calls per minute
        Do: func() (module.Result, error) {
            start := time.Now()
            // ... perform the request ...
            return module.Result{Duration: time.Since(start)}, nil
        },
    },
}
```

A module with no ops is valid — Arbiter will call `Run` and let the module drive its own traffic generation.

### Full example

See [`examples/samplemod`](examples/samplemod) for a working module with args and multiple operations.

## CLI usage

The only supported way to run a test right now is via the `cli` subcommand. Arbiter automatically generates CLI flags for every registered module's args and ops.

```
./my-binary cli [module flags...] [runner flags...]
```

**Module arg flags** are prefixed with the module name:

```
--<module>.<arg-name>   value
```

**Operation flags** are also auto-generated for each op:

```
--<module>.op.<op-name>.rate     uint    # calls per minute (default from Op.Rate)
--<module>.op.<op-name>.disable  bool    # set to true to skip this operation
```

For example, a module named `sample` with an arg `important` and an op `test` produces:

```
--sample.important         int     A very important argument. (required)
--sample.op.test.rate      uint    Rate at which to call the test operation per minute.
--sample.op.test.disable   bool    Disable the test operation.
```

## Runner flags

These flags apply to both the `cli` and `file` subcommands:

| Flag | Short | Default | Description |
|---|---|---|---|
| `--duration` | `-d` | `5m0s` | How long to run the test. Minimum 1 second. |
| `--report-path` | `-r` | `report.yaml` | File path where the YAML report is written. |
| `--interactive` | `-i` | `false` | Show a live TUI with per-operation statistics while the test runs. |

Example:

```
./my-binary cli --duration 2m --report-path results.yaml --sample.important 42
```

Arbiter also stops cleanly on `SIGINT` or `SIGTERM`.

## Report

After a test finishes Arbiter writes a YAML report to the path set by `--report-path`. The report contains timing and success/failure counts per module and operation. The exact schema is subject to change, but a typical report looks like:

```yaml
start: 2024-11-01T10:00:00Z
end: 2024-11-01T10:05:00Z
duration: 5m0s
modules:
  sample:
    operation:
      test:
        executions: 600
        ok: 598
        nok: 2
        timing:
          longest: 15ms
          shortest: 10ms
          average: 11ms
```
