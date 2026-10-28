# Development lifecycle

Rules without a marker are the author's; **(P)** marks a Claude Code proposal awaiting the author's approval. Commands live in `Taskfile.yml` (`task --list`), not here.

## Flow

1. **Intent.** A behavior change starts from a kibi requirement ([kibi.md](kibi.md)); a product or architecture choice starts from a decision in DESIGN.md.
2. **Branch** from `main`; branches are short-lived and `main` is never committed to directly.
3. **Change** the code with tests at the right level ([testing.md](testing.md)) and the code linked to its requirements.
4. **Local pipeline.** Tests, lint and vulnerability checks run on the pre-commit hook, and they are the same checks CI runs, so local green means CI green.
5. **PR** to `main`; CI repeats the checks and runs the end-to-end tests. Merge when green and reviewed.
6. **`main`** deploys to the demo environment.
7. **Release** through a separate release workflow.

## Tooling

- The Taskfile is the single entry point for all tooling, kibi included. Hooks and CI only call task targets, never tools directly.
- Tool versions are pinned in the Taskfile **(P)**; lefthook manages the git hooks **(P)**. Only one tool owns `.git/hooks`.
- `task vuln` fails on any finding, and also when a vulnerability database can't be reached.
- kibi's relationship shards and `symbols.yaml` merge record by record through kibi's merge driver (`.gitattributes`); `task setup` registers it in each clone.

## Commits and PRs

- Conventional Commits; release notes are generated from them **(P)**.
- One concern per PR. The PR says which requirements or decisions it implements and how it was verified.
- Never bypass hooks, and never skip, disable or quarantine tests.

## CI and environments

- CI runs on GitHub Actions **(P)**: the same checks as the hook, end-to-end tests on k3s plus kind (D-12), and multi-arch images with an SBOM, signed with cosign (D-17).
- Demo and test environment: a Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases. It holds synthetic data only: never real broker credentials or portfolio data (GR-12).
- Deploy credentials live in a protected GitHub Environment, never in the repository **(P)**.
- Creating, reaching, deploying to and destroying the demo machine: [demo.md](demo.md).

## Releases

Modelled on OpenClaw: calendar versions, release channels and a separate release workflow.

- Product versions are `YYYY.M.N`, with betas `YYYY.M.N-beta.K`. Tags carry no `v` prefix, because Go would read a root tag `v2026.11.0` as major version 2026 of the module **(P)**.
- Channels: `beta` deploys to demo; `stable` is a GitHub release with signed images, the CRDs and the bootstrap install.
- The release workflow is started manually, does a dry run before it publishes anything, and only the author releases.
- Contracts and SDKs use SemVer, independent of the product version. The Go SDK is tagged `sdk/go/vX.Y.Z`, as Go requires for a submodule **(P)**, and `buf breaking` gates every SDK release **(P)**.

## Dependencies (P)

Pin exact versions and commit lockfiles. Every new dependency needs a reason in its PR. Renovate proposes updates.

## Definition of done

- `task check` is green.
- New behavior has a kibi requirement with its scenario and test linked.
- New domain terms are in [the glossary](../glossary.md).
- New decisions are in DESIGN.md.
