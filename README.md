# OpenFief

**Give your AI agents a fief, not the kingdom.**

Self-hosted portfolio tracker and orchestrator for AI investment agents. Each agent gets a revocable capital lease with hard limits enforced outside the LLM. Bring your own agents, models and brokers.

**Status:** M0 in progress. The domain libraries, the read API, the plugin operator with image signature checks and the egress proxy are built; the core service binary, UI and connector plugins are not yet. The [status page](https://openfief.dev/docs/status/) tracks every decision.

## Documents

- [docs/OVERVIEW.md](docs/OVERVIEW.md): one-page introduction. Start here.
- [docs/DESIGN.md](docs/DESIGN.md): the current, decided design.
- [docs/proposals/](docs/proposals/): what is proposed but not decided, one file per area.
- [docs/research.md](docs/research.md): external research the design draws on.
- [docs/glossary.md](docs/glossary.md): domain terms and abbreviations.
- Requirements live in kibi (`.kb/`); `task kibi:report` renders them as a page.

The same documents, plus concept pages for each part as built and pages generated from the code, are published at **[openfief.dev/docs](https://openfief.dev/docs/)** on every push to `main`. `task docs:build` builds the site locally.

## Development

Install [Task](https://taskfile.dev), run `task setup`, then `task check`. Coding agents start at [AGENTS.md](AGENTS.md); the rules live in [docs/dev/](docs/dev/).

## License

The core is licensed under the [GNU AGPL v3 or later](LICENSE). Plugin contracts and SDKs are Apache-2.0, with their own license file in their directories. Contributions need a [DCO](https://developercertificate.org/) sign-off (`git commit -s`).

> Not investment advice. You run OpenFief on your own accounts, at your own risk.
