@agent/prd.json @agent/progress.txt

You are working on the "dash" project - a Grafana-like monitoring dashboard.

**Tech Stack:**
- Frontend: Vue.js 3 (Composition API + TypeScript)
- Backend: Go API
- Database: PostgreSQL (metadata)
- Data Sources: Prometheus, Loki, Victoria Logs, VictoriaMetrics

## Continuous Development Mode

**Work through multiple features without stopping!** Commit directly to master for each feature (no PRs needed).

## Instructions (Per Feature)

1. **Find next feature:** Pick the highest priority incomplete feature from prd.json

2. **Run tests:**
   - Frontend: `cd frontend && npm run type-check && npm run test`
   - Backend: `cd backend && go test ./...`

3. **Implement the feature** (just this one feature, nothing else)

4. **Update PRD:** Set `passes: true` for completed feature

5. **Update progress.txt:**
```
## Feature N: [Name] - [timestamp]
- What was done:
- Files changed:
- Tests: [passing/failing]
```

6. **Commit and push:**
```bash
git add -A
git commit -m "feat: [description]"
git push origin master
```

7. **Continue:** Move to next feature - DO NOT STOP!

## Only Stop When:

- All features have `passes: true` → output `<promise>COMPLETE</promise>`
- Iteration limit reached (let Ralph script handle this)

## Summary Format (after each feature):

```
✅ Feature N complete: [Name]
   Committed to master
   
🔄 Moving to next feature...
```
