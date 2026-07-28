# Changelog

All notable changes to `flagctl` will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.0.2] - 2026-07-28

### Added
- **`flagctl launch` Subcommand Suite:** Progressive feature launch commands (`add`, `list`, `ramp`, `remove`).
  - `flagctl launch add`: Adds global or cohort-specific launch ramps.
  - `flagctl launch list`: Displays all active launch ramps with index numbers and percentage splits.
  - `flagctl launch ramp`: Ramps launch percentage up or down (`--percent 50`, `--percent 100`).
  - `flagctl launch remove`: Removes a launch ramp by index (`--index N`).
- **`flagctl target` Subcommand Suite:** Ordered rule-based targeting commands (`add`, `list`, `remove`).
  - `flagctl target add`: Adds targeting rules (allowlists, denylists, semver, segment rules).
  - `flagctl target list`: Displays ordered rules + overlap analysis hints.
  - `flagctl target remove`: Removes a specific targeting rule by index (`--index N`).
- **Tier-Based Rule Ordering:** Automatic priority ordering:
  - Tier 1: Denylists (`"off"`, placed at top).
  - Tier 2: Allowlists & Segment rules.
  - Tier 3: Cohort-specific progressive launch ramps.
  - Tier 4: Global fallback launch ramp.
- **Multiple Cohort Launches:** Support for concurrent progressive launches targeting specific user cohorts (e.g. 50% for mobile, 20% for desktop).
- **Explicit Bucket ID Support:** Compliant with `flagd` spec for `{ "var": "<bucketId>" }` (defaults to `userId`, configurable via `--bucket-by`).

### Changed
- Renamed primary progressive rollout command from `flagctl rollout` to `flagctl launch` (`rollout` retained as CLI alias for backward compatibility).
- Updated `flagctl launch ramp` to replace redundant `--complete` flags.

### Removed
- Removed unsafe `flagctl target clear` command to prevent accidental wipes of production targeting rules.

---

## [0.0.1] - 2026-07-28

### Added
- **Initial Release of `flagctl` CLI:** Written in Go (`v1.26.5`).
- **Embedded Schema Registry:** Versioned offline validation for `flagd` `v0` flags and targeting schemas (`schema/v0/flags.json`, `schema/v0/targeting.json`).
- **Workspace Engine:** Automatic root directory and configuration discovery (`.flagctl.json`, `flags.json`, `flags.yaml`).
- **Core CLI Commands:**
  - `flagctl init`: Idempotent workspace initialization.
  - `flagctl create`: Flag creation for `boolean`, `string`, `number`, and `object` types.
  - `flagctl update`: Flag attribute updating with **immutable flag key enforcement**.
  - `flagctl deprecate`: Soft-deprecation keeping `state: ENABLED` to avoid runtime `FLAG_NOT_FOUND` crashes while freezing edits.
  - `flagctl undeprecate`: Un-freezes deprecated flags.
  - `flagctl delete`: Code-aware flag removal with active reference checking.
  - `flagctl validate`: Schema validation against versioned `flagd` JSON schemas.
  - `flagctl list`: Terminal summary table output.
  - `flagctl audit`: Codebase scanner for missing, deprecated, and orphaned flag references (`--strict` mode for CI/CD).
  - `flagctl generate`: Strongly-typed code accessor generation for **TypeScript** (`src/flags.gen.ts`) and **Go** (`pkg/flags/flags.gen.go`).
  - `flagctl version`: Displays CLI version (`0.0.1`).
- **Code-Aware Deletion Guard:** Automatically scans codebase and blocks `flagctl delete` if active code calls exist.
- **Multi-Language Code Scanner:** Audits JS/TS, Go, Python, Java, C#, PHP, and Rust files for OpenFeature SDK flag calls.
