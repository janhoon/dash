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

LLM modules import `github.com/aceobservability/ace/backend/pkg/llm`
(`RegisterLLM`, `AIProvider`). Datasource modules import
`github.com/aceobservability/ace/backend/pkg/datasource`
(`RegisterDatasource`, `Client`). Anthropic is
`github.com/aceobservability/ace-llm-anthropic` (blank-imported from
`internal/handlers`). Prometheus is
`github.com/aceobservability/ace-datasource-prometheus`. Ace registers it at
`init` and injects the SSRF HTTP client. See `docs/adr/0004-module-contracts.md`.
