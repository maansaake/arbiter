---
name: arbiter-module-development
description: 'Write Arbiter load-testing modules for an application. Covers the module interface, how to structure Args and Ops, lifecycle hooks, and the concrete samplemod example.'
argument-hint: 'Optional: the target application or protocol you want the module to exercise'
---

# Arbiter Module Development

## When to Use

- Writing a new Arbiter module to load test an HTTP API, gRPC service, database-backed app, queue consumer, or any other system under test
- Refactoring an existing module to expose better Args or Ops
- Understanding how to implement the `module.Module` interface Arbiter expects
- Looking for a concrete reference implementation in this repo

---

## Core Interface

An Arbiter module is a Go type that implements `pkg/module.Module`:

```go
type Module interface {
    Name() string
    Desc() string
    Args() Args
    Ops() Ops
    Run() error
    Stop() error
}
```

Use it to describe:

- **module identity** via `Name` and `Desc`
- **configuration** via `Args`
- **load-generating operations** via `Ops`
- **setup/teardown** via `Run` and `Stop`

Register one or more modules from your program's `main`:

```go
func main() {
    err := arbiter.Run(module.Modules{mymod.New()}, nil)
    // handle err
}
```

---

## Naming Rules

- `Name()` must be unique and kebab-case / lowercase-alphanumeric-plus-dashes
- Operation names must follow the same pattern
- Do not use reserved module names/prefixes like `arbiter` or `reporter`

The runtime validates module and op names before starting.

---

## Recommended Module Shape

Follow the same pattern as `examples/samplemod/module/module.go`:

```go
type MyModule struct {
    args module.Args
    ops  module.Ops

    baseURL string
    client  *http.Client
}

func New() module.Module {
    m := &MyModule{
        client: &http.Client{},
    }

    m.args = module.Args{
        &module.Arg[string]{
            Name:     "base-url",
            Desc:     "Base URL of the target service.",
            Required: true,
            Value:    &m.baseURL,
        },
    }

    m.ops = module.Ops{
        &module.Op{
            Name: "health",
            Desc: "Calls the health endpoint.",
            Rate: 60,
            Do: func() (module.Result, error) {
                start := time.Now()

                // perform request here

                return module.Result{Duration: time.Since(start)}, nil
            },
        },
    }

    return m
}
```

Keep parsed config and reusable clients on the struct; build `Args` and `Ops` once in `New`.

---

## Args: Exposing Module Configuration

`Args()` returns `module.Args`, which is a list of typed arguments Arbiter exposes for the module.

Supported argument types:

- `int`
- `uint`
- `float64`
- `string`
- `bool`

Use `module.Arg[T]`:

```go
&module.Arg[string]{
    Name:     "host",
    Desc:     "Target host.",
    Required: true,
    Value:    &m.host,
}
```

### Important fields

| Field | Meaning |
|---|---|
| `Name` | Argument name exposed by the module |
| `Desc` | Help text / description |
| `Required` | Startup fails if omitted |
| `Value` | Pointer that receives the parsed value and can also provide a default |
| `Handler` | Optional callback for derived parsing or conversion |
| `Valid` | Optional validator for rejecting invalid values |

### When to use `Value` vs `Handler`

- Use **`Value`** when the parsed type is already the form your module wants
- Use **`Handler`** when you want to derive another field, such as converting milliseconds to `time.Duration`
- You can use both together

Example:

```go
&module.Arg[int]{
    Name: "timeout-ms",
    Desc: "Request timeout in milliseconds.",
    Handler: func(v int) {
        m.timeout = time.Duration(v) * time.Millisecond
    },
}
```

Prefer Args for things that define the workload or target, such as:

- target URL / host / port
- auth token, username, password, API key source
- request payload size
- tenant, topic, queue, table, or endpoint selection
- timeouts and feature toggles

---

## Ops: Exposing Load-Generating Operations

`Ops()` returns `module.Ops`, a list of `*module.Op`. Each op is one independently scheduled workload.

```go
&module.Op{
    Name: "get-user",
    Desc: "Fetch a user by ID.",
    Rate: 120,
    Do: func() (module.Result, error) {
        start := time.Now()

        // execute one unit of work

        return module.Result{Duration: time.Since(start)}, nil
    },
}
```

### Op fields

| Field | Meaning |
|---|---|
| `Name` | Operation name |
| `Desc` | Description of the operation |
| `Disabled` | Skip scheduling this op |
| `Rate` | Target executions **per minute** |
| `Do` | Function Arbiter calls for one execution |

### Important runtime behavior

- Arbiter schedules each op independently
- `Rate` is treated as **calls per minute** by the runtime and generated op settings
- `Rate` must be **greater than zero** for scheduled ops
- `Disabled: true` prevents the op from being scheduled

Each op also gets its own exposed runtime controls through Arbiter:

- per-op **rate override**
- per-op **disable switch**

That means you should define stable, meaningful operations like `login`, `search`, `create-order`, or `publish-message`, and let Arbiter manage their pacing.

### What `Do` should return

- Return `module.Result{Duration: ...}` with the time spent doing the operation
- Return `nil` error on success
- Return a non-nil error for failed attempts; Arbiter records failures in reporting

Keep one `Do` call to one logical unit of work. If you need a multi-step scenario, either:

- keep the full scenario in a single op, or
- model separate behaviors as separate ops with their own rates

---

## Lifecycle: `Run` and `Stop`

Use lifecycle hooks for setup and cleanup:

- `Run()` is called once before traffic starts
- `Stop()` is called once after traffic stops

Good `Run` responsibilities:

- build HTTP/gRPC/database clients
- warm up auth/session state
- create reusable fixtures or IDs
- validate required connectivity early

Good `Stop` responsibilities:

- close clients/connections
- clean up temporary resources created by the module

Avoid heavy setup inside `Do` unless every invocation really needs it.

---

## Practical Module Design Guidance

When building a module for a real application:

1. Put shared clients, config, and reusable state on the module struct.
2. Use `Args` for everything that should vary between environments or test runs.
3. Split `Ops` by user-visible behavior or traffic class, not by tiny implementation detail.
4. Measure the duration around the actual operation work so Arbiter reports useful timings.
5. Return real errors instead of swallowing them; Arbiter uses them for failure counts.
6. Keep operation names stable so reports stay easy to compare over time.

Examples of good op splits:

- API service: `list-products`, `get-product`, `create-cart`, `checkout`
- Messaging system: `publish`, `consume`, `ack`, `redeliver`
- Auth service: `login`, `refresh-token`, `introspect-token`

---

## No-Ops Modules

A module with no ops is valid if you want the module to drive traffic itself from `Run()`.

For normal application load testing, prefer defining explicit `Ops` so Arbiter can schedule, report, and tune them independently.

---

## Concrete Example in This Repo

Use `examples/samplemod` as the main reference:

- `examples/samplemod/main.go` shows how the module is registered with `arbiter.Run`
- `examples/samplemod/module/module.go` shows:
  - a constructor that builds `args` and `ops`
  - one required arg and one handler-based arg
  - multiple operations
  - success and error-returning `Do` functions
  - minimal `Run` / `Stop` implementations

That example is the best template to copy when starting a new module in this repository.
