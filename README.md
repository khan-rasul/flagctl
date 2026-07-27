# `flagctl` — Declarative Feature Flag Management CLI

`flagctl` is a dedicated, open-standard command-line tool for managing **flagd** feature flag configurations (`flags.json` / `flags.yaml`) stored directly in Git version control.

It provides developer-friendly commands to create, rollout, target, update, deprecate, delete, validate, audit, and generate strongly-typed code accessors for your feature flags while guaranteeing strict conformance to the versioned `flagd` schema (`https://flagd.dev/schema/v0/flags.json`).

---

## Key Features & Architecture

* ⚡ **GitOps Native:** Manage flag definitions along with application source code in Git.
* 🔒 **Immutable Flag Keys:** Prevents telemetry fragmentation and runtime errors by keeping flag keys immutable.
* 🛡 **Idempotent Initialization:** Running `flagctl init` safely sets up or updates `.flagctl.json` without overwriting existing flag definitions.
* 📊 **Progressive Rollouts:** Configure percentage splits (`flagctl rollout`) with 100% sum validation and automatic split balancing.
* 🎯 **Attribute & Segment Targeting:** Easily generate JsonLogic rules for targeted user segments (`flagctl target`).
* ❄️ **Safe 2-Stage Deprecation:** `flagctl deprecate` tags metadata and freezes edits while keeping `state: ENABLED` so production applications never suffer `FLAG_NOT_FOUND` runtime crashes.
* 🔍 **Code-Aware Deletion Guard:** `flagctl delete` scans application source code and blocks flag removal if active code references still exist.
* 🩺 **Codebase Auditor:** `flagctl audit` scans source code for missing flags, dead/orphaned flags, and uncleaned deprecated flags.
* 🧬 **Strongly-Typed Accessor Generation:** `flagctl generate` produces type-safe wrapper accessors for TypeScript and Go for 100% compile-time flag safety.
* 🌐 **Offline Schema Registry:** Embedded versioned JSON schemas (`v0`) enable validation without internet access.

---

## Installation

### Via Go Install
```bash
go install github.com/khan-rasul/flagctl@latest
```

### Build from Source
```bash
git clone https://github.com/khan-rasul/flagctl.git
cd flagctl
go build -o flagctl main.go
```

---

## Quickstart

### 1. Initialize Workspace
Run `flagctl init` in your project root:
```bash
$ flagctl init --language typescript

✔ Initialized empty flagd configuration at ./flags.json
```

### 2. Create a Feature Flag
Create a typed feature flag:
```bash
$ flagctl create --key new-checkout-flow --type boolean --default on --description "New checkout UI"

✔ Successfully created flag 'new-checkout-flow' in flags.json
✔ Regenerated typed accessors at src/flags.gen.ts
```

### 3. Configure Progressive Rollout
Set a 50/50 percentage rollout:
```bash
$ flagctl rollout --key new-checkout-flow --splits "on=50,off=50" --bucket-by "userId"

✔ Successfully updated rollout for 'new-checkout-flow' (splits: on=50,off=50, bucketBy: userId)
```

### 4. Target Specific User Segments
Serve variant `"on"` to company employees:
```bash
$ flagctl target --key new-checkout-flow --attribute "email" --operator "endsWith" --value "@company.com" --variant "on"

✔ Successfully added targeting rule for 'new-checkout-flow'
```

### 5. Validate Configuration Against Schema
Validate `flags.json` against embedded `v0/flags.json` schema:
```bash
$ flagctl validate

✔ ./flags.json is valid according to flagd schema v0!
```

### 6. Audit Codebase
Scan application code for missing or orphaned flags:
```bash
$ flagctl audit

🔍 Auditing codebase in /path/to/project against flags.json...
✔ 0 Missing flags (All code calls exist in config)
✔ 0 Deprecated flags called in code
✔ 0 Orphaned flags
```

---

## Command Reference

| Command | Usage | Description |
| :--- | :--- | :--- |
| `flagctl init` | `flagctl init [-f json\|yaml] [-l ts\|go]` | Idempotently initializes `.flagctl.json` and `flags.json`. |
| `flagctl create` | `flagctl create -k key [-t type] [-d default]` | Creates a new flag definition (boolean, string, number, object). |
| `flagctl rollout` | `flagctl rollout -k key -s "on=20,off=80"` | Sets percentage rollout splits with sum validation & auto-balance. |
| `flagctl target` | `flagctl target -k key -a email -v "@test.com"` | Adds JsonLogic attribute-based targeting rule. |
| `flagctl update` | `flagctl update -k key [-s ENABLED\|DISABLED]` | Updates flag state, default variant, or description. |
| `flagctl deprecate` | `flagctl deprecate -k key [-r reason]` | Soft-deprecates and freezes a flag (keeps state `ENABLED`). |
| `flagctl undeprecate` | `flagctl undeprecate -k key` | Un-freezes a deprecated flag. |
| `flagctl delete` | `flagctl delete -k key [-f]` | Code-aware flag deletion (blocks if code calls exist). |
| `flagctl validate` | `flagctl validate [-f flags.json]` | Validates config against versioned `flagd` JSON schema. |
| `flagctl list` | `flagctl list` | Displays terminal summary table of all flags and rollouts. |
| `flagctl audit` | `flagctl audit [--strict]` | Scans codebase for missing, orphaned, or deprecated flags. |
| `flagctl generate` | `flagctl generate [-l ts\|go] [-o path]` | Generates strongly-typed flag accessor helper code. |
| `flagctl version` | `flagctl version` | Outputs `flagctl` CLI version (`0.0.1`). |

---

## GitHub Actions CI/CD Integration

Add `flagctl` to your PR workflow to validate schemas and block PRs with missing/deprecated flag calls:

```yaml
name: Feature Flag CI Check

on:
  pull_request:
    branches: [ main ]

jobs:
  audit-flags:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Install flagctl
        run: go install github.com/khan-rasul/flagctl@latest

      - name: Validate Flag Config Schema
        run: flagctl validate

      - name: Audit Codebase Flags
        run: flagctl audit --strict
```

---

## License

Apache License 2.0
