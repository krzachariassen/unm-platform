# Git Flow Rules

## NEVER commit directly to `main`. All work happens on feature branches.

## Branch Naming

```
<type>/<short-description>

Types: feat, fix, refactor, test, docs, chore
```

Examples:
- `feat/add-interaction-analyzer`
- `fix/capability-visibility-badge`
- `docs/dsl-by-example-guide`
- `refactor/view-presenter-cleanup`

## Single-Task Workflow

### 1. Before Starting Work

```bash
git checkout main
git pull origin main
git checkout -b <type>/<short-description>
```

### 2. During Work

- Commit early and often on the feature branch
- Use conventional commit messages: `<type>(<scope>): <description>`
- Keep commits focused — one logical change per commit

### 3. After Work is Complete

```bash
git add -A
git commit -m "<type>(<scope>): <description>"
git push -u origin HEAD
gh pr create --title "<title>" --body "<summary>"
```

## Parallel Tasks with Git Worktrees

When multiple independent tasks need to run simultaneously, use **git worktrees**
to give each task its own working directory and branch.

### Why Worktrees

A single git checkout can only be on one branch at a time. If two tasks run
in parallel on different branches, they will clobber each other. Worktrees
solve this — each worktree is an independent checkout with its own branch.

### Worktree Directory

All worktrees live under `.worktrees/` at the project root (gitignored).

### Creating a Worktree

```bash
# From the main worktree (project root):
git worktree add .worktrees/<branch-name> -b <type>/<short-description>
```

Example:
```bash
git worktree add .worktrees/feat-empty-state -b feat/empty-state-component
git worktree add .worktrees/chore-remove-uber -b chore/remove-uber-references
```

Each worktree is a full working directory. Navigate to it and work normally:
```bash
cd .worktrees/feat-empty-state
# edit files, run tests, commit — all on feat/empty-state-component branch
```

### Committing and Pushing from a Worktree

```bash
cd .worktrees/<branch-name>
git add -A
git commit -m "<type>(<scope>): <description>"
git push -u origin HEAD
gh pr create --title "<title>" --body "<summary>"
```

### Cleaning Up Worktrees

After a task is complete (PR merged or branch done):
```bash
# From the main worktree (project root):
git worktree remove .worktrees/<branch-name>
git branch -d <type>/<short-description>
```

### Worktree Rules

- **One task = one worktree = one branch = one PR**
- Each worktree has its own working directory — no file conflicts
- Teammates spawned for parallel tasks MUST each get their own worktree
- The lead/orchestrator stays in the main worktree
- Worktree paths use branch name with slashes replaced by dashes:
  `feat/empty-state` → `.worktrees/feat-empty-state`
- Always clean up worktrees after the task is done
- NEVER create worktrees for sequential/dependent tasks — use a single branch

## Rules

- **NEVER** commit to `main` directly
- **NEVER** force-push to `main`
- **ALWAYS** create a feature branch before making any code changes
- **ALWAYS** check current branch before first commit: `git branch --show-current`
  - If on `main`, create a branch FIRST
  - If on another task's branch, create a worktree for the new task
- Branch from `main` (or the current default branch)
- One branch per task/feature — don't mix unrelated changes
- Delete branches and worktrees after merge
- If you receive a NEW task while a previous task is in-progress:
  **DO NOT switch branches**. Create a worktree for the new task instead.

## Agent Protocol

Every agent MUST follow this sequence at the start of any task:

```bash
# 1. Check current branch
current=$(git branch --show-current)

# 2. Determine action
# If on main → create a feature branch
# If on another task's branch → create a worktree for this new task

# Single task (on main):
git checkout -b <type>/<description>

# New task while another is in-progress (not on main):
git worktree add .worktrees/<branch-slug> -b <type>/<description> main
cd .worktrees/<branch-slug>

# 3. ... do work ...

# 4. Commit, push, PR
git add -A
git commit -m "<type>(<scope>): <description>"
git push -u origin HEAD
```
