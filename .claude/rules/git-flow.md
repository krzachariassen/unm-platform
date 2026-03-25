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

## Workflow

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
git push -u origin HEAD
```

Then create a PR (or tell the user the branch is ready for review).

## Rules

- **NEVER** commit to `main` directly
- **NEVER** force-push to `main`
- **ALWAYS** create a feature branch before making any code changes
- **ALWAYS** check current branch before first commit: `git branch --show-current`
  - If on `main`, create a branch FIRST
- Branch from `main` (or the current default branch)
- One branch per task/feature — don't mix unrelated changes
- Delete branches after merge

## Agent Protocol

Every agent MUST follow this sequence at the start of any task:

```bash
# 1. Check current branch
git branch --show-current

# 2. If on main, create a feature branch
git checkout -b <type>/<description>

# 3. ... do work ...

# 4. Commit on the feature branch
git add -A
git commit -m "<type>(<scope>): <description>"

# 5. Push
git push -u origin HEAD
```
