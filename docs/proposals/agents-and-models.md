# Proposals: agents and models (M2)

Not decided. When the author decides an item, it moves into [DESIGN.md](../DESIGN.md) as a decision and leaves this file.

This area covers running agents (GR-5, GR-7, GR-9, GR-10), the model-portfolio maintainer (GR-4) and model choice per role (GR-8, GR-11). The execution path the agents feed is in [control-plane.md](control-plane.md).

## Agents

- **[P-29] Agents out of process, intents instead of orders.** Agents run out of process and sandboxed, and see only data scoped to their lease. Agent and orchestrator talk over A2A; the tools OpenFief exposes are served over MCP. Agents send order intents with a thesis and invalidation conditions, never raw broker orders: GR-7's "tools to make transactions" is the gateway's intent API, gated by a capability token (see T-5).
- **[P-30] Agents reach the world only through OpenFief.** Agent sandboxes have no direct network egress. Models are reached through the model router, which enforces P-33's tiers and meters model spend per agent; news and filings arrive through the reader (P-31); trades go through the gateway. The M0 egress proxy and per-plugin NetworkPolicy (D-11, D-32) are the mechanism. Without this, P-31 and P-33 are advice to agent authors rather than controls.
- **[P-31] Reader and decider.** Defend against prompt injection by splitting a reader (untrusted text in, typed and validated research notes out) from a decider that sees only those notes. Every input carries its source, and a thesis changes only on two independent sources (see R-6).
- **[P-32] Maintainer without trading tools.** The GR-4 maintainer has no transaction tools. OpenFief computes the figures (weights, drift); the model narrates them and proposes model-portfolio changes; the owner applies them. Open: OQ-12.

## Models and data

- **[P-33] Data tiers and the model router.** A model router maps each role to allowed models and each model to allowed data tiers, denying by default. Tier rules are defaults the owner can loosen, with an explicit warning, so the router never overrides GR-11.

  | Tier | Contents | Default allowed models |
  |---|---|---|
  | T0 | Whole-portfolio holdings and amounts, transactions, tax data | Local models only |
  | TF | One fief's own lease, cash, positions and working orders | The owning agent's configured models, cloud included |
  | T1 | Thesis text and instrument list | Per owner configuration |
  | T2 | Public data: prices, filings, news | Per owner configuration |

  TF exists because a cloud-hosted thesis agent (GR-8) has to see its own budget to size orders (T-6).

## Contracts

Interface outlines, not code. Every plugin also follows the decided plugin rules (D-11, D-16, D-17, D-27, D-32).

- **[P-34] Agent contract.**

  | Aspect | Outline |
  |---|---|
  | Purpose | A thesis agent product that OpenFief runs (GR-9, GR-10). |
  | Manifest | Publisher, version, roles it needs (for example reader, decider) with model requirements, data tiers requested (T1, T2, TF), events it subscribes to, tools it needs, sandbox resources, guarantee mode (gateway-mediated or broker-native). |
  | Transport | A2A to the orchestrator; MCP tools served by OpenFief. |
  | Inbound | Wake with an event and a context: thesis, own-fief snapshot, typed research notes, recent decisions. Never T0. |
  | Tools | Read fief state. Read market data. Search news and filings through the reader. Register or clear price triggers. Submit or cancel an intent. Propose a thesis update. Request a lease change (always needs approval). Call models through the router. |
  | Outbound | Intents (instrument, side, size, limit price, time in force, thesis reference, invalidation conditions, source references), research notes, a status summary for the owner. |
  | Broker-native agents | Such as IBKR's: they skip the intent path, run with guarantee level "broker wall only" on a dedicated linked account, and OpenFief observes them through a read connector (P-19, P-20). |

- **[P-35] Model provider contract.**

  | Aspect | Outline |
  |---|---|
  | Purpose | Supply models to roles through the router (GR-11). |
  | Manifest | Provider, location (local, or cloud with region), declared data retention and training use, models with capabilities (tool use, context window, structured output), pricing. |
  | Operations | List models. Generate (chat with tools, structured output). Embeddings (optional). Report usage for metering. |
  | Constraints | Receives only the tiers its binding allows. Never receives credentials. Every call is logged with role, tier, model and hashes (P-18). |

## Orchestration

- **[P-36] Event catalogue.** Wakes and control-plane events, delivered through the event store (P-23).

  | Event | Emitted by | Delivered to | Effect |
  |---|---|---|---|
  | Scheduled wake | Orchestrator | The agent | Agent runs with its current context |
  | Price trigger | Orchestrator, from market data | The agent | Agent wakes on a level, move or invalidation price it registered |
  | News trigger | Reader | The agent's decider | Agent wakes with a typed research note, never raw text |
  | Fill | Gateway | Agent, ledger, audit log, UI | Ledger moves notional from reserved to committed; reconciliation runs |
  | Lease breach | Ledger, reconciler or drawdown breaker | Owner, gateway, agent | Fief suspended; working orders cancelled |
  | Kill | Owner or drawdown breaker | Gateway, orchestrator, agents in scope | Tokens revoked; working orders cancelled; fiefs suspended |
  | Approval decided | Approval service | Agent, gateway, audit log | Intent routed or reservation released |
  | Policy denied | Policy engine | Agent, audit log | Intent rejected; the agent may revise it |

- **[P-37] Agent-side entities.**

  | Entity | Key attributes |
  |---|---|
  | Agent | Plugin (publisher, version), model per role, data tiers granted, subscribed events, guarantee level; works on one thesis and holds at most one active fief |
  | Thesis | Statement, instruments, horizon (long-term or short-term), invalidation conditions, sources, version history |
  | ResearchNote | Typed claim from the reader, with source tags and the number of independent sources |
  | ModelPortfolio | Versioned target weights; changes proposed by the maintainer and applied by the owner |

## Security model

### Model-portfolio maintainer

- **Data seen:** T0 (full holdings, the model portfolio and its targets) and T2.
- **Models:** configurable (GR-4, GR-11); the author would use only a local model (GR-4).
- **Threats:** T0 sent to a cloud provider through misconfiguration; prompt injection through public data skewing summaries; the maintainer silently changing the model portfolio; invented numbers in summaries.
- **Controls:** the router pins the role to local providers (P-33); no transaction tools and precomputed figures (P-32); every call logged (P-18).

### Thesis agent: research

- **Data seen:** T1 (its own thesis and instruments), T2, and TF (its own fief). Never T0, never another fief.
- **Models:** per role, owner-configured (GR-11); cloud frontier models allowed (GR-8).
- **Threats:** prompt injection through news and filings, including adversarial news aimed at multi-agent trading systems (R-6); the thesis revealed to a cloud provider; a malicious or compromised agent plugin; runaway model spend.
- **Controls:** P-29, P-30, P-31 and P-33; research notes in the audit log (P-18).

## Milestone

- **[P-38] M2 acceptance criteria: agents on paper.**
  1. The default agent (OQ-4) runs one thesis end to end on paper: it wakes on schedule, price and news events and submits intents with thesis and invalidation conditions.
  2. A second agent runs under the same contract without changes to OpenFief.
  3. The maintainer runs on a local model, produces a summary and a proposed model-portfolio change, and has no transaction tool.
  4. A thesis agent runs on a cloud model. A test that routes T0 data to a cloud-bound role is blocked and audited.
  5. On a prompt-injection test set, injected text never reaches a decider as raw text and no thesis changes on a single source. Where an injected intent is forced through, all M1 criteria (P-27) still hold.
  6. The UI shows each agent's guarantee level.
  7. Each agent's paper track record (returns, drawdown, invalidation conditions hit) is recorded as input to P-21.

## Tensions

- **[T-6] Data tiers vs GR-8 and GR-11.** Keeping all holdings and amounts on local models conflicts with a cloud-hosted thesis agent (GR-8), which needs at least its own budget and positions to size an order; and a hard tier rule would limit GR-11's "the user chooses which model sees what". Proposed: the TF tier and overridable defaults in P-33. See OQ-14.

## Open questions

- **[OQ-4]** Default agent: the author's or pfranczyk's? Should the agent contract be co-designed with pfranczyk?
- **[OQ-6]** Which local model runs the GR-4 maintainer role on the author's hardware? (Depends on OQ-18.)
- **[OQ-12]** What authority does the GR-4 maintainer have? Does "maintains" include placing trades (for example rebalancing the defensive part), or only maintaining the model portfolio? Who sets the size of the offensive part: the owner, the model portfolio, or the maintainer? (P-32)
- **[OQ-14]** May a cloud-hosted thesis agent see its own fief's amounts and positions? Are tier rules hard limits or defaults the owner can override under GR-11? (T-6, P-33)
- **[OQ-18]** What hardware is available for local models?
