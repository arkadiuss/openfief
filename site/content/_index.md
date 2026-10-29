---
title: OpenFief
description: A self-hosted, open-source portfolio tracker that grows into an orchestrator for AI investment agents. Each agent gets a revocable slice of capital with limits enforced outside the model.
eyebrow: Open source · self-hosted · Kubernetes
lede: OpenFief tracks everything you own across brokers, then lets AI agents trade a slice of it. Each agent gets a revocable capital lease, a fief, with limits enforced by code and separate broker accounts, never by a prompt.
features:
  - title: Limits live outside the model
    icon: fief
    text: An order gateway, a budget ledger, policy rules and a separate broker account enforce each lease. A bad prompt or a poisoned news article can't talk its way past code.
  - title: Your cluster, your data
    icon: cluster
    text: Runs in your own Kubernetes cluster behind one Service, which you expose however you like, through your Gateway or an Ingress. Your full holdings stay on your hardware; cloud models see only what you allow.
  - title: Every plugin holds a fief too
    icon: wall
    text: Each connector and view runs in its own pod, with its own identity, only the Secrets it was granted, no network path except to the core, and outbound HTTPS only to allow-listed hosts through an egress proxy.
  - title: Signed or it doesn't run
    icon: seal
    text: The operator deploys a plugin image only when a trusted signer signed its digest, checked offline against Sigstore's trust root. No setting deploys an unsigned image.
  - title: A ledger that keeps its history
    icon: ledger
    text: Double-entry and append-only. Re-imports change nothing, corrections add a version instead of overwriting, and balance assertions catch import mistakes.
  - title: Polish investors first
    icon: bond
    text: Interactive Brokers, XTB statements and treasury bonds valued from their terms and official inflation data, with Polish tax rules in a swappable jurisdiction pack.
stage: "OpenFief is in its first milestone, M0: the tracker, with Polish broker and treasury bond imports. The agents come in later milestones, and nothing here is ready for real money yet."
milestones:
  - name: M0
    title: Tracker
    text: Imports, the ledger, valuations, the read API, the built-in UI and isolated plugins.
    now: true
  - name: M1
    title: Control plane
    text: Leases, the order gateway and approvals, on paper trading.
  - name: M2
    title: Agents
    text: Thesis agents as plugins, each with its own fief and models.
  - name: M3
    title: A live fief
    text: The first small lease trading real money.
help: The design is public, with every decision numbered and every proposal open to critique. The most useful input now is a hard look at [the security model](../../docs/DESIGN.md#37-security-tracking-and-imports), and experience with the Interactive Brokers API or its linked accounts as a Polish resident. Start with the [introduction](../../docs/OVERVIEW.md), then [how it works](../../docs/architecture.md).
---

## How it works

Statements, prices and rates come in through connectors. The core keeps the record: an append-only ledger and a valuation of every holding on every day. Views show it, and your own scripts read it through the same read-only API. Each plugin runs in its own pod with only the permissions you granted, and reaches the internet only through an egress proxy that knows which hosts it may call.

```mermaid
flowchart LR
  inst@{ icon: "fief:institution", form: "square", label: "Brokers, NBP, GUS,<br/>market data", pos: "b", h: 40 }
  subgraph cluster["Your Kubernetes cluster"]
    direction LR
    egress@{ icon: "fief:egress", form: "square", label: "Egress proxy", pos: "b", h: 40 }
    conn@{ icon: "fief:connector", form: "square", label: "Connectors", pos: "b", h: 40 }
    core@{ icon: "fief:gate", form: "square", label: "Core: ledger,<br/>valuations, read API", pos: "b", h: 40 }
    db@{ icon: "fief:database", form: "square", label: "PostgreSQL", pos: "b", h: 40 }
    views@{ icon: "fief:view", form: "square", label: "Views: built-in UI,<br/>Ghostfolio", pos: "b", h: 40 }
  end
  you@{ icon: "fief:owner", form: "square", label: "You", pos: "b", h: 40 }
  scripts@{ icon: "fief:script", form: "square", label: "Your scripts", pos: "b", h: 40 }
  inst -- "statements,<br/>prices" --> egress --> conn
  conn -- "transactions" --> core
  core --- db
  core -- "read API" --> views
  views -- "pages" --> you
  core -- "read API" --> scripts
```

[See how the parts fit together](../../docs/architecture.md), or [take the tour](../../docs/quickstart.md) of a running OpenFief.

## Build on it

OpenFief is a platform before it is an app: everything outside the core is a plugin. Connectors bring data in, views show it, and agents, later, trade inside their fief. A plugin is a container image whose signed manifest says what it needs: which hosts it calls, which Secrets it reads, what it may do. You approve exactly that, and the operator enforces it.

```mermaid
flowchart LR
  code@{ icon: "fief:script", form: "square", label: "Write it<br/>on the Go SDK", pos: "b", h: 40 }
  image@{ icon: "fief:seal", form: "square", label: "Sign the image<br/>and its manifest", pos: "b", h: 40 }
  owner@{ icon: "fief:owner", form: "square", label: "The owner approves<br/>its grant", pos: "b", h: 40 }
  op@{ icon: "fief:operator", form: "square", label: "The operator runs it,<br/>isolated", pos: "b", h: 40 }
  core@{ icon: "fief:gate", form: "square", label: "The core calls it", pos: "b", h: 40 }
  code --> image --> owner --> op --> core
```

[Build a plugin](../../docs/guides/build-a-plugin.md) walks through a real connector, from its manifest to its signature.

## Where it's going

The tracker comes first; the agents follow on top of it. This is the planned flow for one agent:

1. **Spawn a thesis agent.** "Company X is undervalued; three-year horizon." Give it a fief of 5,000 PLN. It sees its thesis, public data and its own fief, nothing else.
2. **It proposes a trade:** a limit order, its reasoning, and what would prove the thesis wrong. The gateway checks it against the lease and your rules.
3. **You approve it on your phone.** The order goes to a separate broker account that holds only capital set aside for agents.
4. **Pull the kill switch** if something looks off. The agent loses its trading rights, its open orders are cancelled, and the rest of your portfolio is untouched.

```mermaid
flowchart LR
  A@{ icon: "fief:agent", form: "square", label: "Thesis agent<br/>in its own pod", pos: "b", h: 40 }
  G@{ icon: "fief:gate", form: "square", label: "Order gateway:<br/>lease and rules", pos: "b", h: 40 }
  P@{ icon: "fief:phone", form: "square", label: "Your phone", pos: "b", h: 40 }
  B@{ icon: "fief:institution", form: "square", label: "Separate<br/>broker account", pos: "b", h: 40 }
  K@{ icon: "fief:killswitch", form: "square", label: "Kill switch", pos: "b", h: 40 }
  A -- "trade proposal" --> G
  G -- "approval request" --> P
  P -- "approve" --> G
  G -- "order" --> B
  K -. "revoke and cancel" .-> G
```

## Why a fief

In medieval Europe, a lord granted a vassal a fief: land to run as the vassal saw fit, held on conditions, and taken back when they broke. The vassal ruled the fief, never the kingdom. OpenFief makes the same deal with software agents, and writes the conditions in code.

Our mascot keeps the gate. Kasztan is a żubr, the bison of Białowieża, a royal forest whose bison Polish kings put under crown protection in the 16th century: a creature living inside a guarded domain, under rules it can't argue with. His name is the *kasztelan*, the castellan who kept a castle for the crown, and *kasztan*, the chestnut of his coat.
