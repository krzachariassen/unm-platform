# Validation Pipeline

Run the full validation suite deterministically. Every gate below MUST be
executed — do not skip any gate regardless of earlier results. Collect all
results and report at the end.

## Gate Execution

Run these gates in order. Capture the exit code and output of each.

### Gate 1: Backend Build

```bash
cd backend && go build ./cmd/server/ 2>&1
echo "EXIT_CODE: $?"
cd backend && go build ./cmd/cli/ 2>&1
echo "EXIT_CODE: $?"
```

### Gate 2: Backend Vet

```bash
cd backend && go vet ./... 2>&1
echo "EXIT_CODE: $?"
```

### Gate 3: Backend Tests

```bash
cd backend && go test ./... 2>&1
echo "EXIT_CODE: $?"
```

### Gate 4: Frontend Type Check

```bash
cd frontend && npx tsc --noEmit 2>&1
echo "EXIT_CODE: $?"
```

### Gate 5: Frontend Build

```bash
cd frontend && npx vite build 2>&1
echo "EXIT_CODE: $?"
```

### Gate 6: Security Scan

Scan for hardcoded secrets, API keys, and PII patterns in staged or recently
modified files. Check for:

- Hardcoded API keys (`sk-`, `AKIA`, `ghp_`, `Bearer `)
- Hardcoded secrets in variable assignments (`password =`, `secret =`, `token =`
  with string literals)
- PII in log statements (`log.*email`, `log.*password`, `log.*ssn`)
- `.env` files with real values committed (not `.env.example`)

```bash
# Check for common secret patterns in Go and TypeScript files
rg -i '(sk-[a-zA-Z0-9]{20,}|AKIA[A-Z0-9]{16}|ghp_[a-zA-Z0-9]{36}|password\s*[:=]\s*"[^"]+"|secret\s*[:=]\s*"[^"]+"|api_key\s*[:=]\s*"[^"]+")' backend/ frontend/ --type-add 'code:*.go' --type-add 'code:*.ts' --type-add 'code:*.tsx' --type code 2>&1 || true
echo "EXIT_CODE: $?"
```

Report any matches as FAIL with the file:line and matched pattern.
If no matches, report PASS.

### Gate 7: Diff Coverage (Advisory)

If changes have been made (unstaged or staged files exist):

```bash
# Identify changed Go files
git diff --name-only HEAD -- 'backend/*.go' 2>/dev/null | head -20

# Run coverage for changed packages
cd backend && go test -coverprofile=coverage.out ./... 2>&1
cd backend && go tool cover -func=coverage.out 2>&1 | tail -1
echo "EXIT_CODE: $?"
```

Report total coverage percentage. This gate is advisory (WARN not FAIL) in
the current phase — it reports coverage but does not block.

## Report Format

After running ALL gates, produce this exact report format:

```
╔══════════════════════════════════════════════════╗
║              VALIDATION REPORT                    ║
╠══════════════════════════════════════════════════╣
║ Gate 1: Backend Build        [PASS|FAIL]          ║
║ Gate 2: Backend Vet          [PASS|FAIL]          ║
║ Gate 3: Backend Tests        [PASS|FAIL]          ║
║ Gate 4: Frontend TypeCheck   [PASS|FAIL]          ║
║ Gate 5: Frontend Build       [PASS|FAIL]          ║
║ Gate 6: Security Scan        [PASS|FAIL]          ║
║ Gate 7: Diff Coverage        [PASS|WARN|N/A]      ║
╠══════════════════════════════════════════════════╣
║ Result: [ALL GATES PASSED | X GATE(S) FAILED]    ║
╚══════════════════════════════════════════════════╝
```

For any FAIL gate, include the exact error output below the report:

```
### Gate X: [Gate Name] — FAILED
[exact error output, first 50 lines]
```

## Self-Correction Protocol

If any gate FAILS and you are running as part of the orchestrator workflow:

1. Report the failure details to the orchestrating agent
2. The orchestrator gets ONE retry: fix the specific failure and re-run
   ONLY the failed gate(s)
3. If the retry also fails: stop and escalate to the human with the full
   failure report
4. Do NOT retry more than once — repeated failures indicate a real problem

## Usage

```
/validate              # Run all gates
/validate backend      # Run gates 1-3 + 6 only
/validate frontend     # Run gates 4-5 + 6 only
```

$ARGUMENTS
