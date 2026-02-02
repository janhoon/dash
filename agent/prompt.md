@agent/prd.json @agent/progress.txt

You are working on the "dash" project - a Grafana-like monitoring dashboard.

**Tech Stack:**
- Frontend: Vue.js 3 (Composition API + TypeScript)
- Backend: Go API
- Database: PostgreSQL (metadata)
- Data Source: Prometheus

## Instructions

1. Find the highest priority feature to work on and work only on that feature. YOU decide
which feature to work on that has not been completed - it might not be the first feature.

2. Run type checks and tests:
   - Frontend: `cd frontend && npm run type-check && npm run test`
   - Backend: `cd backend && go test ./...`
   - Make sure they are passing

3. Update the PRD with the work that was done (set passes: true when complete)

4. Append the progress.txt file with the progress that was made. This is for the next person working on the codebase.
Example output for progress.txt:

```
## [Feature Name] - [timestamp]
- What was done:
- Files changed:
- Blockers/notes for next iteration:
```

5. Make a git commit of this feature

6. Push the branch to GitHub:
   - Create a feature branch if not already on one
   - Use descriptive branch name: feat/feature-name
   - Push: `git push origin HEAD`

7. Create a Pull Request on GitHub:
   - Use `gh pr create` with appropriate title and description
   - Describe what was implemented and what testing is needed
   - Reference the feature from the PRD
   - Include any setup/testing instructions
   - Capture the PR URL from the output

ONLY WORK ON A SINGLE FEATURE!

After creating the PR, output:

<promise>PR_CREATED</promise>

This signals that the feature is complete and ready for review. The loop will stop here.

DO NOT continue to the next feature. DO NOT output <promise>COMPLETE</promise> unless ALL features in the PRD have passes: true.

## When ALL Features are Complete

If all features in prd.json have "passes": true, output:

<promise>COMPLETE</promise>
