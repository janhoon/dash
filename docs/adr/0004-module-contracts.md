---
status: accepted
---

# Module contracts for LLM and datasource adapters

> **Implementation status:** Accepted. Contracts live in `backend/pkg/llm` and
> `backend/pkg/datasource`. Anthropic is the first out-of-tree LLM
> (`ace-llm-anthropic`, [#448](https://github.com/aceobservability/ace/issues/448)).
> OpenAI-compat (`openai`, `openrouter`, `ollama`, `custom`) is the second
> (`ace-llm-openai-compat`, [#450](https://github.com/aceobservability/ace/issues/450)).
> Copilot is the third (`ace-llm-copilot`, [#450](https://github.com/aceobservability/ace/issues/450)).
> Prometheus is the first out-of-tree datasource (`ace-datasource-prometheus`,
> [#449](https://github.com/aceobservability/ace/issues/449)). ClickHouse is
> also out-of-tree (`ace-datasource-clickhouse`,
> [#451](https://github.com/aceobservability/ace/issues/451),
> [#461](https://github.com/aceobservability/ace/pull/461)). VictoriaMetrics
> and VictoriaLogs are also out-of-tree (`ace-datasource-victoriametrics`,
> `ace-datasource-victorialogs`,
> [#451](https://github.com/aceobservability/ace/issues/451)). Loki is also
> out-of-tree (`ace-datasource-loki`,
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

## Host wiring

LLM adapters register from module `init`. Ace blank-imports
`github.com/aceobservability/ace-llm-anthropic`, and that module's `init` calls
`RegisterLLM("anthropic", New)`. OpenAI-compat is the second. Ace blank-imports
`github.com/aceobservability/ace-llm-openai-compat`, and that module's `init`
calls `RegisterLLM` for `openai`, `openrouter`, `ollama`, and `custom`. Copilot
is the third. Ace blank-imports `github.com/aceobservability/ace-llm-copilot`,
and that module's `init` calls `RegisterLLM("copilot", New)`.
Remaining datasource factories live in
`internal/datasource` and call `RegisterDatasource` from `init`. Prometheus,
VictoriaMetrics, Loki, and VictoriaLogs are out-of-tree datasources.
`ace-datasource-prometheus`, `ace-datasource-victoriametrics`,
`ace-datasource-loki`, and
`ace-datasource-victorialogs` implement the contract. Ace registers type
`prometheus` from `internal/datasource/register.go`, type `victoriametrics`
from `internal/datasource/register_victoriametrics.go`, type `loki` from
`internal/datasource/register_loki.go`, and type
`victorialogs` from `internal/datasource/register_victorialogs.go` via the
typed `register()` helper, and injects `ssrf.DatasourceClient` (auth-wrapped)
into `New(url, httpClient)`. ClickHouse is also out-of-tree
(`ace-datasource-clickhouse`). Ace registers type `clickhouse` from
`internal/datasource/register_clickhouse.go` and injects
`ssrf.DatasourceClient` (auth-wrapped) plus `AuthConfig` (database) into
`New(url, httpClient, authConfig)`. Ace owns `TestConnection` for those types via
`runHTTPConnectionCheck` and `HTTPClient()` (`testRegisteredHTTPConnection`).
It does not call module `connect.go`. HTTP
handlers call `internal/datasource.NewClient` and `llm.New` /
`llm.RequireKnown`. `datasource.NewClient` looks up `cfg.Type`. There is no
parallel type argument.

`vmalert` and `alertmanager` are not query `Client` types. `TestConnection`
keeps a dedicated path for them.

An LLM module must depend on a `ace/backend` version that does not import that
module. Ace's current module then depends on the adapter. That avoids a Go
module import cycle. Do not add a nested `pkg/llm` module for this.

## Out of scope

WASM, go-plugin, or marketplace loading.
