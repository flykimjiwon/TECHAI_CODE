# Git Workflow Guide

## Branch Naming

```
feature/user-authentication
fix/login-redirect-loop
hotfix/critical-security-patch
refactor/extract-payment-service
docs/update-api-reference
test/add-integration-tests
```

**Pattern**: `<type>/<short-description>`

## Commit Convention

```
feat: add user registration endpoint
fix: prevent race condition in order processing
docs: update deployment guide
refactor: extract validation logic to middleware
test: add unit tests for discount calculator
chore: upgrade dependencies
perf: optimize database queries with indexing
style: format code with prettier
```

**Format**: `<type>: <imperative verb> <what changed>`

## PR Template

```markdown
## Summary
Brief description of changes (1-2 sentences)

## Changes
- Added user authentication with JWT
- Refactored error handling middleware
- Updated API documentation

## Test Plan
- [ ] Unit tests pass
- [ ] Manual test: login flow
- [ ] Edge case: expired token handling

## Screenshots (if UI change)
```

## Rebase vs Merge

**Rebase**: feature branch를 main에 rebase
```bash
git checkout feature/new-api
git rebase main
git push --force-with-lease
```

**Merge**: PR 머지 시 squash merge (깔끔한 히스토리)

## Hotfix Workflow

```bash
git checkout main
git pull
git checkout -b hotfix/critical-bug
# fix + commit
git push -u origin hotfix/critical-bug
# PR → main
git checkout develop
git cherry-pick <hotfix-commit-sha>
```

## Stash Workflow

```bash
git stash push -m "WIP: user profile feature"
git checkout main
# work on hotfix
git checkout feature/profile
git stash pop
```

**패턴**: feature branch → rebase main → squash merge. Hotfix는 cherry-pick으로 develop 동기화.
