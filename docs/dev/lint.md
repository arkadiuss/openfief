# Lint

Strict on purpose: lint is how agents get fast, unambiguous feedback. Rules without a marker are the author's; **(P)** marks a Claude Code proposal pending approval. The Go config is `.golangci.yml` (run it with `task lint`); configs for other languages come with their code.

## Go (golangci-lint v2) (P)

- **Linters on by default.** `default: all`. Each disabled linter is listed in the config with a one-line reason.
- **Must stay on:**
  - `errcheck`, `govet` (all analyzers), `staticcheck`, `gosec`, `errorlint`, `wrapcheck`;
  - `exhaustive`, `nilerr`, `bodyclose`, `contextcheck`, `noctx`;
  - `gocritic`, `revive`, `gocognit`, `cyclop`, `funlen`, `paralleltest`, `thelper`, `tparallel`, `usetesting`;
  - `gochecknoglobals`, `nolintlint`, `depguard`, `forbidigo`.
- **Formatters:** `gofumpt` and `goimports`.
- **`depguard`** keeps domain packages pure ([code-organization.md](code-organization.md)): no `context`, `database/sql`, `log`, `math/rand`, `net`, `os`, `runtime`, `sync`, `syscall` or `unsafe`. depguard matches import prefixes only, so "no context imports another's adapters" can't be expressed there; that needs go-arch-lint once code exists **(P)**.
- **`forbidigo`** bans `time.Now`, `time.Since`, `time.Until`, `fmt.Print*` and `panic` in domain packages. `float64` for money isn't expressible in a linter; reviews catch it.
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
