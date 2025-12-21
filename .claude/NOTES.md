# Development Notes

## Development Servers
Keep both servers running during development sessions.

### Go Server (port 8080)
```bash
# Start
cd /Users/mike/dev/bouncebot/server && ./server

# Kill
pkill -f "./server"

# Rebuild and restart
cd /Users/mike/dev/bouncebot/server && go build -o server . && ./server
```

### Vue Dev Server (port 5173)
```bash
# Start (from client/vue1)
npm run dev

# Kill
lsof -ti:5173 | xargs kill -9

# If port 5173 is in use by old server, kill it first
```

## Docker (for deployment only)
Docker is for CI/CD and deployment, not local development. Use native binaries locally.

## Workflow

- Always submit new changes via PRs (not direct pushes to main)
- Always update progress documentation before making PRs:
  - `client/vue1/PROGRESS.md` for implementation progress
  - `planning/REFACTORING_PLAN.md` for refactoring tasks
- When merging PRs, use `gh pr merge --merge --delete-branch` to clean up remote branches
