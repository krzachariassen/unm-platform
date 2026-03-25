# Code Review Template

## Review: [PR/Change Title]

### Summary
[1-2 sentence summary of what was changed and why]

### Architecture Check
- [ ] Clean Architecture layers respected
- [ ] No import violations
- [ ] Business logic in correct layer
- [ ] API contract matches frontend types

### Test Check
- [ ] New code has tests
- [ ] Tests written before implementation (TDD)
- [ ] Edge cases covered
- [ ] AI tests use real API (no mocking)

### Code Quality
- [ ] Error handling correct
- [ ] Naming follows conventions
- [ ] No dead code or unused variables
- [ ] No god functions (>50 lines should be decomposed)

### Security
- [ ] No hardcoded secrets
- [ ] Input validation present
- [ ] No unsafe type assertions without checks

### Findings

#### Critical (must fix)
1. **[File:Line]** -- [Description]

#### Important (should fix)
1. **[File:Line]** -- [Description]

#### Minor (nit)
1. **[File:Line]** -- [Description]

### Verdict
- [ ] Approve
- [ ] Approve with minor changes
- [ ] Request changes (critical findings)
