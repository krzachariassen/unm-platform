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

### Gate 4: Frontend Build (includes TypeScript check)

The frontend `npm run build` script runs `tsc -b && vite build`. This is the
ONLY command to use — it matches exactly what the Dockerfile and CI run.
Do NOT use `tsc --noEmit` as it has different behavior than `tsc -b`.

```bash
cd frontend && npm run build 2>&1
echo "EXIT_CODE: $?"
```

### Gate 5: Security Scan

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

### Gate 6: CI Parity — Git-Tracked Test Fixtures

Tests that reference files via relative paths (e.g. `examples/`, `testdata/`)
will pass locally if the file exists on disk — even if it is gitignored and
therefore absent in CI's clean checkout. This gate catches that class of bug.

```bash
# Find all fixture file paths referenced in Go test files,
# resolve them relative to each test file's location,
# and verify each resolved path is tracked in git.
echo "=== Checking test fixture files are git-tracked ==="
python3 - <<'PYEOF'
import os, subprocess, sys, re

repo_root = subprocess.check_output(["git", "rev-parse", "--show-toplevel"]).decode().strip()
failures = []

for dirpath, _, files in os.walk(os.path.join(repo_root, "backend")):
    for fname in files:
        if not fname.endswith("_test.go"):
            continue
        filepath = os.path.join(dirpath, fname)
        with open(filepath) as f:
            content = f.read()
        # Extract quoted paths that look like fixture references
        for match in re.finditer(r'"([^"]*(?:examples|testdata)[^"]*\.(?:yaml|json|unm))"', content):
            ref = match.group(1)
            resolved = os.path.normpath(os.path.join(dirpath, ref))
            rel = os.path.relpath(resolved, repo_root)
            result = subprocess.run(
                ["git", "ls-files", "--error-unmatch", rel],
                cwd=repo_root, capture_output=True
            )
            if result.returncode != 0:
                failures.append(f"  NOT IN GIT: {rel}  (referenced in {os.path.relpath(filepath, repo_root)})")

if failures:
    print("FAIL — fixture files referenced in tests but not committed to git:")
    for f in sorted(set(failures)):
        print(f)
    sys.exit(1)
else:
    print("PASS — all test fixture files are git-tracked.")
    sys.exit(0)
PYEOF
echo "EXIT_CODE: $?"
```

### Gate 7: CI Parity — No Private Registry in package-lock.json

`npm ci` uses resolved URLs from `package-lock.json` directly, ignoring
`.npmrc` overrides. If the lockfile was generated on a machine with a private
registry configured (e.g. Artifactory, Verdaccio, internal mirrors), CI runners
without credentials will fail with E401. This gate catches that.

```bash
echo "=== Checking package-lock.json for private registry URLs ==="
PRIVATE_HITS=$(grep -c '"resolved"' frontend/package-lock.json 2>/dev/null && \
  grep '"resolved"' frontend/package-lock.json | \
  grep -v '"https://registry\.npmjs\.org/' | \
  grep -v '"https://registry\.yarnpkg\.com/' | \
  head -5)
if [ -n "$PRIVATE_HITS" ]; then
  echo "FAIL — package-lock.json contains non-public resolved URLs:"
  echo "$PRIVATE_HITS"
  echo "Fix: rm frontend/package-lock.json frontend/node_modules && cd frontend && npm install --registry https://registry.npmjs.org/ --userconfig /dev/null"
  echo "EXIT_CODE: 1"
else
  echo "All resolved URLs point to public registry."
  echo "EXIT_CODE: 0"
fi
```

### Gate 8: Diff Coverage (Advisory)

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
║ Gate 4: Frontend Build       [PASS|FAIL]          ║
║ Gate 5: Security Scan        [PASS|FAIL]          ║
║ Gate 6: CI Parity — Fixtures [PASS|FAIL]          ║
║ Gate 7: CI Parity — Registry [PASS|FAIL]          ║
║ Gate 8: Diff Coverage        [PASS|WARN|N/A]      ║
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
/validate backend      # Run gates 1-3 + 5-6 only
/validate frontend     # Run gates 4-5 + 7 only
```

$ARGUMENTS
