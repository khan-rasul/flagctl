# `flagctl` — Declarative Feature Flag Management CLI

`flagctl` is a dedicated, open-standard command-line tool for managing **flagd** feature flag configurations (`flags.json` / `flags.yaml`) stored directly in Git version control.

It provides developer-friendly commands to create, launch, target, update, deprecate, delete, validate, audit, and generate strongly-typed code accessors for your feature flags while guaranteeing strict conformance to the versioned `flagd` schema (`https://flagd.dev/schema/v0/flags.json`).

---

## Key Features & Architecture

* ⚡ **GitOps Native:** Manage flag definitions along with application source code in Git.
* 🔒 **Immutable Flag Keys:** Prevents telemetry fragmentation and runtime errors by keeping flag keys immutable.
* 🚀 **Progressive Feature Launches (`flagctl launch`):** Ramp feature percentage up/down (`flagctl launch ramp --percent 50`), configure multi-variant splits, or create cohort-specific progressive launches. *(Alias: `rollout`)*.
* 🎯 **Ordered Rule Targeting (`flagctl target`):** Configure Allowlists, Denylists, SemVer version gates, and segment rules with automatic Tier-Based priority ordering.
* 🪣 **Spec-Compliant Bucketing:** Explicit Bucket ID support (`{ "var": "<bucketId>" }`, defaulting to `userId`, customizable via `--bucket-by`).
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

### 3. Add Rule Targeting (Allowlist / Denylist)
Add an Allowlist rule for company employees and a Denylist rule for competitors:
```bash
# Allowlist rule (company employees get 'on')
$ flagctl target add --key new-checkout-flow --attribute "email" --operator "endsWith" --value "@company.com" --variant "on"

# Denylist rule (competitors get 'off', inserted at top of rule chain)
$ flagctl target add --key new-checkout-flow --attribute "email" --operator "endsWith" --value "@competitor.com" --variant "off" --top
```

### 4. Progressive Feature Launch & Ramping
Start a 20% canary launch, then ramp it up to 50%:
```bash
# Start canary launch at 20%
$ flagctl launch add --key new-checkout-flow --percent 20 --variant "on"

# Ramp launch up to 50%
$ flagctl launch ramp --key new-checkout-flow --percent 50
```

### 5. Inspect Active Rules & Launches
View ordered rule chain and overlap analysis:
```bash
$ flagctl target list --key new-checkout-flow

ORDERED TARGETING RULES FOR 'new-checkout-flow' (Default: on):
  [1] DENYLIST : email endsWith "@competitor.com" -> off
  [2] ALLOWLIST: email endsWith "@company.com"    -> on
  [3] LAUNCH   : fractional rollout (fractional)
```

### 6. Validate Configuration Against Schema
Validate `flags.json` against embedded `v0/flags.json` schema:
```bash
$ flagctl validate

✔ ./flags.json is valid according to flagd schema v0!
```

---

## Command Reference

### Launch Commands (`flagctl launch` / `rollout`)
| Command | Usage | Description |
| :--- | :--- | :--- |
| `flagctl launch add` | `flagctl launch add -k key -p 20 -v on` | Adds global or cohort-specific launch ramp. |
| `flagctl launch list` | `flagctl launch list -k key` | Lists active launches with index numbers & percentages. |
| `flagctl launch ramp` | `flagctl launch ramp -k key -p 50` | Ramps launch percentage up/down (0% to 100%). |
| `flagctl launch remove` | `flagctl launch remove -k key -i 1` | Removes a launch ramp by index. |

### Targeting Commands (`flagctl target`)
| Command | Usage | Description |
| :--- | :--- | :--- |
| `flagctl target add` | `flagctl target add -k key -a email -v "@test.com" --variant on` | Adds targeting rule (allowlist, denylist, semver, segment). |
| `flagctl target list` | `flagctl target list -k key` | Displays ordered rule chain + overlap hints. |
| `flagctl target remove` | `flagctl target remove -k key -i 1` | Removes a targeting rule by index. |

### Core Management Commands
| Command | Usage | Description |
| :--- | :--- | :--- |
| `flagctl init` | `flagctl init [-f json\|yaml] [-l ts\|go]` | Idempotently initializes `.flagctl.json` and `flags.json`. |
| `flagctl create` | `flagctl create -k key [-t type] [-d default]` | Creates a new flag definition (boolean, string, number, object). |
| `flagctl update` | `flagctl update -k key [-s ENABLED\|DISABLED]` | Updates flag state, default variant, or description. |
| `flagctl deprecate` | `flagctl deprecate -k key [-r reason]` | Soft-deprecates and freezes a flag (keeps state `ENABLED`). |
| `flagctl undeprecate` | `flagctl undeprecate -k key` | Un-freezes a deprecated flag. |
| `flagctl delete` | `flagctl delete -k key [-f]` | Code-aware flag deletion (blocks if code calls exist). |
| `flagctl validate` | `flagctl validate [-f flags.json]` | Validates config against versioned `flagd` JSON schema. |
| `flagctl list` | `flagctl list` | Displays terminal summary table of all flags and rollouts. |
| `flagctl audit` | `flagctl audit [--strict]` | Scans codebase for missing, orphaned, or deprecated flags. |
| `flagctl generate` | `flagctl generate [-l ts\|go] [-o path]` | Generates strongly-typed code accessor helper code. |
| `flagctl version` | `flagctl version` | Outputs `flagctl` CLI version (`0.0.1`). |

---

## License

Apache License 2.0
