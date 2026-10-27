# Development lifecycle

Rules without a marker are the author's; **(P)** marks a Claude Code proposal pending approval (some are already implemented; implemented doesn't mean approved). The Taskfile and hooks exist (D-41). CI and the release workflow aren't written yet.

## Flow

1. **Intent.** A behavior change starts from a kibi requirement ([kibi.md](kibi.md)). Product or architecture choices start from a decision in DESIGN.md (`design-docs` skill).
2. **Branch** from `main`. Branches are short-lived; `main` is never committed to directly.
3. **Change.** Code, tests at the right level ([testing.md](testing.md)), code linked to requirements.
4. **Local pipeline.** `task check` runs on the pre-commit hook. It's the same command CI runs, so local green means CI green.
5. **PR** to `main`; CI repeats the checks and runs e2e. Merge when green and reviewed.
6. **`main`** deploys to the demo environment automatically.
7. **Release** through the separate release workflow.

## Commands

All entry points are [Taskfile](https://taskfile.dev) targets (D-41). Hooks and CI call targets, never tools directly. First time in a clone: install Task, then run `task setup`.

| Target | Does | Where |
|---|---|---|
| `task setup` | Installs the pinned tools into `.bin/`, kibi (`npm ci`) and the git hooks | once per clone |
| `task check` | fmt check, lint, `govulncheck` + `npm audit`, tests, docs provenance check, full `kibi check` | local, CI |
| `task check:commit` | The same, but with kibi checking staged changes only | pre-commit hook |
| `task test` | Unit and replay-integration tests (offline) | local, CI |
| `task e2e` | Full flow on a local k3s; refuses non-local clusters | local, CI |
| `task fmt` / `task lint` / `task vuln` / `task docs:check` | The individual steps | local |
| `task kibi:*` | `check`, `staged`, `sync`, `status`, `doctor`, `search -- <topic>`, `report` | local, hooks, CI |

Go steps skip cleanly until a `go.mod` exists. Tool versions are pinned in `Taskfile.yml`.

**Hooks.** lefthook (`lefthook.yml`) only calls task targets:
- `pre-commit`: `task check:commit`.
- `post-checkout` (branch switch), `post-merge`, `post-rewrite` (rebase): `task kibi:sync`.

kibi's own hooks stay uninstalled (`kibi init --no-hooks`); its hook commands live in `kibi:staged` and `kibi:sync`. Never let two tools own `.git/hooks`.

**Offline (P).** If a vulnerability database is unreachable, `task vuln` warns locally and fails in CI (`CI=true`). Findings always fail.

## Commits and PRs

- **Conventional Commits** (`feat:`, `fix:`, `docs:` ...) **(P)**. Release notes are generated from them.
- **One concern per PR.** The PR says which requirements or decisions it implements and how it was verified.
- **Never bypass hooks** (`--no-verify`). Never skip, disable or quarantine tests.

## CI (GitHub Actions) (P)

**On every PR and on `main`:**
1. `task check`.
2. `task e2e` on k3s in the runner, plus kind for the portability matrix (D-12).
3. Build images: multi-arch, with an SBOM.
4. Sign them with cosign, keyless (D-17).

**On `main` only:** deploy to demo, then run a smoke subset of e2e against it.

## Environments

| Environment | Where | Data |
|---|---|---|
| Local | k3s or k3d on the developer's machine | Synthetic only |
| Demo / test | A Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases | Synthetic only: no real broker credentials or portfolio data, ever (GR-12) |

Deploy credentials live in a protected GitHub Environment, not in the repo **(P)**.

## Releases

Modelled on OpenClaw: calendar versions, release channels, and a separate release workflow.

**Product versions:** `YYYY.M.N` (e.g. `2026.11.0`), with betas `YYYY.M.N-beta.K`. Tags carry **no `v` prefix** **(P)**, because Go would read a root tag `v2026.11.0` as major version 2026 of the Go module.

**Channels:**
- `beta`: deploys to demo.
- `stable`: a GitHub Release with signed images, CRDs and the bootstrap chart.

**Release workflow:** separate, started manually (`workflow_dispatch`).
1. `preflight`: a dry run. Builds, verifies, drafts notes, publishes nothing.
2. `publish`: tags, pushes signed images, creates the release.
3. Only the author releases.

**Contracts and SDKs:** SemVer, independent of the product version.
- The protobuf package version (`v1`) is the wire contract.
- The Go SDK is tagged `sdk/go/vX.Y.Z`, as Go requires for a submodule.
- The TypeScript SDK follows npm SemVer.
- `buf breaking` gates every SDK release.

## Dependencies (P)

Pin exact versions and commit lockfiles. Every new dependency needs a reason in the PR (supply chain). Renovate proposes updates; CI's `govulncheck` and `npm audit` must pass.

## Definition of done

- `task check` is green and `kibi check` passes.
- New behavior has a requirement with its scenario and test linked.
- New terms are in the ubiquitous language.
- Decisions are recorded in DESIGN.md.
