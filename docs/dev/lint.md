# Lint

Strict on purpose: lint is how agents get fast, unambiguous feedback. Rules without a marker are the author's; **(P)** marks a Claude Code proposal pending approval. Tool configs aren't written yet; this is their contract.

## Go (golangci-lint v2) (P)

- **Linters on by default.** `default: all`. Each disabled linter is listed in the config with a one-line reason.
- **Must stay on:**
  - `errcheck`, `govet` (all analyzers), `staticcheck`, `gosec`, `errorlint`, `wrapcheck`;
  - `exhaustive`, `nilerr`, `bodyclose`, `contextcheck`, `noctx`;
  - `gocritic`, `revive`, `gocognit`, `cyclop`, `funlen`, `paralleltest`, `thelper`, `tparallel`, `usetesting`;
  - `gochecknoglobals`, `nolintlint`, `depguard`, `forbidigo`.
- **Formatters:** `gofumpt` and `goimports`.
- **`depguard`** enforces the layering in [code-organization.md](code-organization.md): domain packages import only the standard library, `internal/kernel` and approved pure libraries; no context imports another's adapters.
- **`forbidigo`** bans in domain packages: `time.Now`, `math/rand`, `fmt.Print*`, `log.*`, `os.Getenv`, and `float64` for money fields (where expressible).
- **`nolintlint`:** every `//nolint` names the linter and gives a reason (`//nolint:gosec // reason`). Bare `//nolint` fails.

## Other languages (P)

| Area | Tool | Notes |
|---|---|---|
| Protobuf contracts | `buf lint`, `buf breaking` | `buf breaking` against the last release enforces the SDK's SemVer promise |
| TypeScript SDK | Biome (strict) + `tsc --strict` | No `any`, no non-null assertions |
| Kubernetes manifests | `kubeconform` + `kube-linter` | CRDs and the bootstrap chart |
| Shell | `shellcheck` | |
| Dependencies | `govulncheck`, `npm audit` | Part of `task check` |

## Rules for agents

- Never weaken lint to get green: no new `//nolint`, disabled linter, raised threshold or excluded path without the author's explicit OK in the PR.
- Fix the cause. If a rule is wrong for this codebase, propose the config change separately, with the reason.
- Generated code (`gen/`) is excluded by path, never by `//nolint`.
