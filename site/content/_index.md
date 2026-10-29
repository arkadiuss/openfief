---
title: OpenFief
description: A self-hosted, open-source portfolio tracker that grows into an orchestrator for AI investment agents. Each agent gets a revocable slice of capital with limits enforced outside the model.
eyebrow: Open source · self-hosted · early
lede: OpenFief tracks everything you own across brokers, then lets AI agents trade a slice of it. Each agent gets a revocable capital lease, a fief, with limits enforced by code and separate broker accounts, never by a prompt.
stage: "Milestone M0 in progress: a tracker with Polish broker and treasury bond imports. Agents come later. Not ready for real money yet."
features:
  - title: Limits live outside the model
    state: planned
    text: An order gateway, a budget ledger, policy rules and a separate broker account enforce each lease. A bad prompt or a poisoned news article can't talk its way past code.
  - title: Your cluster, your data
    state: partly built
    text: Runs in your own Kubernetes cluster, reached through `kubectl port-forward`. Your full holdings stay on your hardware; cloud models see only what you allow.
  - title: Every plugin holds a fief too
    state: built
    text: Each connector and view runs in its own pod, with its own identity, only the Secrets it was granted, no network path except to the core, and outbound HTTPS only to allow-listed hosts through an egress proxy.
  - title: Signed or it doesn't run
    state: built
    text: The operator deploys a plugin image only when a trusted signer signed its digest, checked offline against Sigstore's trust root. No setting deploys an unsigned image.
  - title: A ledger that keeps its history
    state: built
    text: Double-entry and append-only. Re-imports change nothing, corrections add a version instead of overwriting, and balance assertions catch import mistakes.
  - title: Polish investors first
    state: partly built
    text: Interactive Brokers, XTB statements and treasury bonds valued from their terms and official inflation data, with Polish tax rules in a swappable jurisdiction pack.
---

## Where it's going

OpenFief is built in milestones. The tracker comes first; the agents follow on top of it. This is the planned flow for one agent:

1. **Spawn a thesis agent.** "Company X is undervalued; three-year horizon." Give it a fief of 5,000 PLN. It sees its thesis, public data and its own fief, nothing else.
2. **It proposes a trade:** a limit order, its reasoning, and what would prove the thesis wrong. The gateway checks it against the lease and your rules.
3. **You approve it on your phone.** The order goes to a separate broker account that holds only capital set aside for agents.
4. **Pull the kill switch** if something looks off. The agent loses its trading rights, its open orders are cancelled, and the rest of your portfolio is untouched.

```mermaid
flowchart LR
  A["Thesis agent<br/>in its own pod"] -- "trade proposal" --> G["OpenFief gateway<br/>lease and rules check"]
  G -- "approval request" --> P["Your phone"]
  P -- "approve" --> G
  G -- "order" --> B["Separate broker account"]
  K["Kill switch"] -. "revoke and cancel" .-> G
```

## Why a fief

In medieval Europe, a lord granted a vassal a fief: land to run as the vassal saw fit, held on conditions, and taken back when they broke. The vassal ruled the fief, never the kingdom. OpenFief makes the same deal with software agents, and writes the conditions in code.

Our mascot keeps the gate. Kasztan is a żubr, the bison of Białowieża, a royal forest whose bison Polish kings put under crown protection in the 16th century: a creature living inside a guarded domain, under rules it can't argue with. His name is the *kasztelan*, the castellan who kept a castle for the crown, and *kasztan*, the chestnut of his coat.

## Help shape it

The design is public, with every decision numbered and every proposal open to critique. The most useful input now is a hard look at [the security model](../../docs/DESIGN.md), and experience with the Interactive Brokers API or its linked accounts as a Polish resident. Start with the [overview](../../docs/OVERVIEW.md), then the [architecture](../../docs/architecture.md).
