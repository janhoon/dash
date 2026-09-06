---
status: accepted
---

# Module contracts for LLM and datasource adapters

> **Implementation status:** Accepted. Contracts live in `backend/pkg/llm` and
> `backend/pkg/datasource`. Prometheus is extracted to
> `ace-datasource-prometheus` ([#449](https://github.com/aceobservability/ace/issues/449)).
> Remaining `ace-llm-*` / `ace-datasource-*` extracts follow
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

LLM factories still live in `internal/handlers` and call `llm.RegisterLLM` from
`init`. Remaining datasource factories live in `internal/datasource` and call
`RegisterDatasource` from `init`. Prometheus is the first out-of-tree adapter:
`ace-datasource-prometheus` implements the contract; Ace's `init` registers
type `prometheus` and injects `ssrf.DatasourceClient` (auth-wrapped) into
`New(url, httpClient)`. HTTP handlers call `internal/datasource.NewClient` and
`llm.New` / `llm.RequireKnown`. `datasource.NewClient` looks up `cfg.Type`.
There is no parallel type argument.

`vmalert` and `alertmanager` are not query `Client` types. `TestConnection`
keeps a dedicated path for them.

## Out of scope

Moving Anthropic or remaining adapters out of tree. WASM, go-plugin, or
marketplace loading.
