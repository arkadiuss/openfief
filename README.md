# OpenFief

**Give your AI agents a fief, not the kingdom.**

Self-hosted portfolio tracker and orchestrator for AI investment agents. Each agent gets a revocable capital lease with hard limits enforced outside the LLM. Bring your own agents, models and brokers.

**Status:** Design phase, no code yet.

## Documents

- [docs/OVERVIEW.md](docs/OVERVIEW.md): one-page introduction. Start here.
- [docs/DESIGN.md](docs/DESIGN.md): ground rules, research findings, proposals, tensions and open questions, each tagged by source.
- [docs/SPEC.md](docs/SPEC.md): requirements, domain model, lifecycles, plugin contracts, events and milestones derived from the design.

## Development

Install [Task](https://taskfile.dev), run `task setup`, then `task check`. Coding agents start at [AGENTS.md](AGENTS.md); the rules live in [docs/dev/](docs/dev/).

> Not investment advice. You run OpenFief on your own accounts, at your own risk.
