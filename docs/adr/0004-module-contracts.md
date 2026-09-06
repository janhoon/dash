---
status: accepted
---

# Module contracts for LLM and datasource adapters

> **Implementation status:** Accepted. Contracts live in `backend/pkg/llm` and
> `backend/pkg/datasource`. Implementations live only in out-of-tree modules.
> Ace has no in-tree LLM provider or datasource query-client implementations.
>
> LLM modules: `ace-llm-anthropic` ([#448](https://github.com/aceobservability/ace/issues/448)),
> `ace-llm-openai-compat` (`openai`, `openrouter`, `ollama`, `custom`),
> `ace-llm-copilot` ([#450](https://github.com/aceobservability/ace/issues/450)).
>
> Datasource modules: `ace-datasource-prometheus` ([#449](https://github.com/aceobservability/ace/issues/449)),
> `ace-datasource-clickhouse`, `ace-datasource-victoriametrics`,
> `ace-datasource-victorialogs`, `ace-datasource-loki`,
> `ace-datasource-tempo`, `ace-datasource-victoriatraces` (shared tracing
> helpers in `ace-datasource-tempo/tracing`),
> `ace-datasource-elasticsearch`, `ace-datasource-cloudwatch`,
> `ace-datasource-vmalert`, `ace-datasource-alertmanager`
> ([#451](https://github.com/aceobservability/ace/issues/451)).
>
> Parent epic: [#446](https://github.com/aceobservability/ace/issues/446).

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

Ace blank-imports LLM modules so each module `init` calls `RegisterLLM`:

- `ace-llm-anthropic` registers `anthropic`
- `ace-llm-openai-compat` registers `openai`, `openrouter`, `ollama`, `custom`
- `ace-llm-copilot` registers `copilot`

Ace does not construct those providers itself. HTTP handlers call `llm.New`
and `llm.RequireKnown`.

Ace registers query datasource types from `internal/datasource` via the typed
`register()` helper and injects `ssrf.DatasourceClient` through
`newDatasourceHTTPClient`:

| Type | Register file | Module `New` |
| --- | --- | --- |
| `prometheus` | `register.go` | `New(url, httpClient)` |
| `victoriametrics` | `register_victoriametrics.go` | `New(url, httpClient)` |
| `loki` | `register_loki.go` | `New(url, httpClient)` |
| `victorialogs` | `register_victorialogs.go` | `New(url, httpClient)` |
| `tempo` | `register_tempo.go` | `New(url, httpClient)` |
| `victoriatraces` | `register_victoriatraces.go` | `New(url, httpClient)` |
| `clickhouse` | `register_clickhouse.go` | `New(url, httpClient, authConfig)` |
| `elasticsearch` | `register_elasticsearch.go` | `New(url, authConfig, httpClient)` |
| `cloudwatch` | `register_cloudwatch.go` | `New(cfg, httpClient)` with a bare `ssrf.DatasourceClient` (no auth wrap: SDK SigV4 on dual-host metrics/logs) |

HTTP handlers call `internal/datasource.NewClient`. That looks up `cfg.Type`.
There is no parallel type argument.

Ace owns `TestConnection` for prometheus, victoriametrics, loki, victorialogs,
clickhouse, tempo, victoriatraces, and elasticsearch via
`runHTTPConnectionCheck` and `HTTPClient()` (`testRegisteredHTTPConnection`).
It does not call those modules' `connect.go`. CloudWatch uses the module
`TestConnection`.

`vmalert` and `alertmanager` are not query `Client` types. Ace does not
`RegisterDatasource` for them. `ace-datasource-vmalert` and
`ace-datasource-alertmanager` implement `New(url, httpClient)`. Ace injects
`ssrf.DatasourceClient` (auth-wrapped) through thin factories. `TestConnection`
keeps a dedicated Ace path that calls `runHTTPConnectionCheck`
(`IsLocalURL` / `ValidateDatasourceURL`). Do not use the module `connect.go`
copy on Ace's path.

An LLM module must depend on a `ace/backend` version that does not import that
module. Ace's current module then depends on the adapter. That avoids a Go
module import cycle. Do not add a nested `pkg/llm` module for this.

## Out of scope

WASM, go-plugin, or marketplace loading.
