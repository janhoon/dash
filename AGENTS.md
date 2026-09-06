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

LLM modules import `github.com/aceobservability/ace/backend/pkg/llm`
(`RegisterLLM`, `AIProvider`). Datasource modules import
`github.com/aceobservability/ace/backend/pkg/datasource`
(`RegisterDatasource`, `Client`). See `docs/adr/0004-module-contracts.md`.

Anthropic is the first out-of-tree LLM. Ace blank-imports
`github.com/aceobservability/ace-llm-anthropic` so the module `init` calls
`RegisterLLM`. OpenAI-compat (`openai`, `openrouter`, `ollama`, `custom`) is
the second. Ace blank-imports
`github.com/aceobservability/ace-llm-openai-compat` so the module `init`
registers those four types. Copilot is the third. Ace blank-imports
`github.com/aceobservability/ace-llm-copilot` so the module `init`
calls `RegisterLLM("copilot", New)`. Prometheus is the first out-of-tree datasource
(`github.com/aceobservability/ace-datasource-prometheus`). Ace registers it
at `init` and injects the SSRF HTTP client. VictoriaMetrics is also
out-of-tree (`github.com/aceobservability/ace-datasource-victoriametrics`).
Type `victoriametrics` is registered from
`backend/internal/datasource/register_victoriametrics.go` via the typed
`register()` helper. Ace owns `TestConnection` through
`runHTTPConnectionCheck`. It does not call module `connect.go`.
