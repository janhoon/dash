---
title: "Contributing"
---

# Ace

## Design System

Always read DESIGN.md before making any visual or UI decisions.
All font choices, colors, spacing, and aesthetic direction are defined there.
Do not deviate without explicit user approval.
In QA mode, flag any code that doesn't match DESIGN.md.

## Git

Prefer worktrees for any medium or larger tasks (tasks that would affect 5 or more files)

## Agent skills

### Issue tracker

Issues are tracked in the `aceobservability/ace` GitHub repo via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical triage roles mapped 1:1 to default label strings. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.

### Module contracts

LLM and datasource implementations live only in out-of-tree modules. Ace has
no in-tree provider or query-client implementations.

Ace keeps contracts, host wiring, and SSRF / `TestConnection` barriers:

- LLM contract: `github.com/aceobservability/ace/backend/pkg/llm`
  (`RegisterLLM`, `AIProvider`). Construction is `llm.New`.
- Datasource contract: `github.com/aceobservability/ace/backend/pkg/datasource`
  (`RegisterDatasource`, `Client`). Construction is `datasource.NewClient`.
- LLM blank-imports so module `init` calls `RegisterLLM`.
- Datasource `register_*.go` files inject `newDatasourceHTTPClient` into module
  `New(...)`.
- Ace-owned connection tests: `runHTTPConnectionCheck`, `ssrf.IsLocalURL`,
  `HTTPClient()`. Do not call module `connect.go` on Ace's path.

See `docs/adr/0004-module-contracts.md`.

LLM modules Ace blank-imports:

- `ace-llm-anthropic` (`anthropic`)
- `ace-llm-openai-compat` (`openai`, `openrouter`, `ollama`, `custom`)
- `ace-llm-copilot` (`copilot`)

Datasource query types Ace registers from `internal/datasource` (prometheus
from `register.go`; others from `register_<type>.go`):

- prometheus, victoriametrics, loki, victorialogs, tempo, victoriatraces:
  `New(url, httpClient)`
- clickhouse: `New(url, httpClient, authConfig)`
- elasticsearch: `New(url, authConfig, httpClient)` (`authConfig` carries
  index/field settings)
- cloudwatch: `New(cfg, httpClient)` with a bare `ssrf.DatasourceClient` (no
  auth wrap: SDK SigV4, dual-host metrics/logs)

Shared tracing helpers live in `ace-datasource-tempo/tracing`.

Ace owns `TestConnection` for prometheus, victoriametrics, clickhouse, loki,
victorialogs, tempo, victoriatraces, and elasticsearch through
`testRegisteredHTTPConnection` / `runHTTPConnectionCheck` and `HTTPClient()`.
CloudWatch `TestConnection` uses the module method.

VMAlert (`ace-datasource-vmalert`) and Alertmanager
(`ace-datasource-alertmanager`) are out-of-tree but not query `Client` types.
Ace does not `RegisterDatasource` for them. Ace injects the same SSRF HTTP
client into `New(url, httpClient)`. `TestConnection` stays on Ace's path
(`runHTTPConnectionCheck` with `IsLocalURL` / `ValidateDatasourceURL`).
Do not call module `connect.go` / module `TestConnection`.
