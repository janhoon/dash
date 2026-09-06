---
status: accepted
---

# Module contracts for LLM and datasource adapters

> **Implementation status:** Accepted. Contracts live in `backend/pkg/llm` and
> `backend/pkg/datasource`. Anthropic is the first out-of-tree LLM
> (`ace-llm-anthropic`, [#448](https://github.com/aceobservability/ace/issues/448)).
> Prometheus is the first out-of-tree datasource (`ace-datasource-prometheus`,
> [#449](https://github.com/aceobservability/ace/issues/449)). VictoriaMetrics
> is also out-of-tree (`ace-datasource-victoriametrics`,
> [#451](https://github.com/aceobservability/ace/issues/451)). Remaining
> `ace-llm-*` / `ace-datasource-*` extracts follow
> [#446](https://github.com/aceobservability/ace/issues/446).

Ace keeps LLM and datasource **contracts plus registries** in this repo.
Adapters register from `init`. Construction is a registry lookup and fails
closed on an unknown type. There is no plugin SDK repo until a second consumer
needs one. There is no runtime loader.

## Import paths

| Module family | Import | Register | Construct |
| --- | --- | --- | --- |
| LLM | `github.com/aceobservability/ace/backend/pkg/llm` | `RegisterLLM` | `llm.New` |
| Datasource | `github.com/aceobservability/ace/backend/pkg/datasource` | `RegisterDatasource` | `datasource.NewClient` (lookup `cfg.Type`) |

Do not import `internal/handlers` from an LLM module. Do not import
`internal/datasource` from a datasource module.

`internal/` is not a stable path for other Go modules. `pkg/` is.

## In-tree wiring

Remaining LLM factories live in `internal/handlers` and call `llm.RegisterLLM`
from `init`. Anthropic is the first out-of-tree LLM. Ace blank-imports
`github.com/aceobservability/ace-llm-anthropic`, and that module's `init` calls
`RegisterLLM("anthropic", New)`. Remaining datasource factories live in
`internal/datasource` and call `RegisterDatasource` from `init`. Prometheus and
VictoriaMetrics are out-of-tree datasources. `ace-datasource-prometheus` and
`ace-datasource-victoriametrics` implement the contract. Ace registers type
`prometheus` from `internal/datasource/register.go` and type `victoriametrics`
from `internal/datasource/register_victoriametrics.go` via the typed
`register()` helper, and injects `ssrf.DatasourceClient` (auth-wrapped) into
`New(url, httpClient)`. Ace owns `TestConnection` for those types via
`runHTTPConnectionCheck`. It does not call module `connect.go`. HTTP
handlers call `internal/datasource.NewClient` and `llm.New` /
`llm.RequireKnown`. `datasource.NewClient` looks up `cfg.Type`. There is no
parallel type argument.

`vmalert` and `alertmanager` are not query `Client` types. `TestConnection`
keeps a dedicated path for them.

An LLM module must depend on a `ace/backend` version that does not import that
module. Ace's current module then depends on the adapter. That avoids a Go
module import cycle. Do not add a nested `pkg/llm` module for this.

## Out of scope

Moving remaining adapters out of tree. WASM, go-plugin, or marketplace loading.
