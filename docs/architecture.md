---
title: Architecture
description: How OpenFief's parts fit together in M0, what is built today, and which page explains each part.
covers:
  - cmd/**
  - deploy/**
---

OpenFief runs in the owner's own Kubernetes cluster. A core written in Go owns the record of every holding and transaction. Everything that talks to the outside world is a plugin: its own container, with only the permissions the owner granted it. An operator turns plugin resources into running, isolated workloads, and an egress proxy is the only way out to the internet. This page shows the whole design for M0 (the first milestone, a tracker with Polish imports) and marks what is built today.

## The system and its neighbors

The owner reaches OpenFief through `kubectl port-forward` by default (D-28), and installs plugins with the `openfief` command-line tool using their own kubeconfig (D-46). OpenFief reaches out only through the egress proxy, and only to hosts a plugin's grant allows (D-32).

```mermaid
flowchart LR
  owner(["Owner"])
  subgraph cluster["Owner's Kubernetes cluster"]
    of["OpenFief<br/>namespace openfief"]
  end
  ibkr["IBKR Flex Web Service"]
  nbp["NBP<br/>exchange rates"]
  gus["GUS<br/>inflation data"]
  md["EODHD<br/>end-of-day prices"]
  gf["Ghostfolio<br/>owner's own install"]
  reg["Container registry<br/>signed images"]
  owner -- "browser via port-forward,<br/>openfief CLI" --> of
  of -- "egress proxy,<br/>granted hosts only" --> ibkr
  of --> nbp
  of --> gus
  of --> md
  of --> gf
  reg -- "digest-pinned images" --> of
```

## Parts, as designed and as built

Green parts are built, amber parts are partly built, and gray parts are decided but not built yet. The table below says what exists in each.

```mermaid
flowchart TB
  classDef built fill:#d8efd9,stroke:#2f7d32,color:#1d1d1f
  classDef partial fill:#fff1c7,stroke:#a77d00,color:#1d1d1f
  classDef missing fill:#f3f3f3,stroke:#9a9a9a,color:#555,stroke-dasharray: 4 3
  cli(["openfief CLI"]):::missing
  subgraph ns["namespace openfief"]
    op["Operator<br/>Plugin + PluginGrant"]:::partial
    proxy["Egress proxy<br/>Smokescreen"]:::built
    core["Core service<br/>ledger, valuation, read API;<br/>connector calls not built"]:::partial
    pg[("PostgreSQL")]:::partial
    subgraph plugins["Plugins, one pod each"]
      ui["Built-in UI view"]:::missing
      gfv["Ghostfolio view"]:::missing
      ibkrc["IBKR Flex connector"]:::partial
      xtb["XTB export parser"]:::missing
      bonds["Treasury bond connector"]:::partial
      mdc["Market-data connector"]:::missing
      nbpc["NBP FX"]:::partial
    end
  end
  sdk["Go SDK<br/>connector contract"]:::built
  sdk -. "plugins build on" .-> plugins
  cli -- "applies resources" --> op
  op -- "creates and constrains" --> plugins
  op -- "allow-lists" --> proxy
  core --> pg
  core -- "calls connectors" --> plugins
  plugins -- "outbound HTTPS" --> proxy
```

| Part | Decided in | Built today | Explained in |
|---|---|---|---|
| Operator | D-13, D-15 to D-17, D-42 to D-45, D-56, D-57 | `cmd/operator` and the reconciler in `internal/plugins/operator`, plus install manifests in `deploy/`. It deploys an image only when a trusted signer signed its digest, checked offline with sigstore-go against a bundled Sigstore trust root (`internal/plugins/cosign`; `cmd/sigstore-trusted-root` refreshes the root). The trusted signers are OpenFief's CI, the publishers pinned in the plugin catalog (`internal/plugins/catalog`, empty today) and any the owner adds in a ConfigMap | [Operator](concepts/operator.md) |
| Egress proxy | D-32, D-63 | `cmd/egress-proxy`, embedding Smokescreen through `BuildProxy`. It checks each plugin's `openfief-egress` token with a TokenReview, caches accepted reviews, and allows only the hosts the grant names | [Egress](concepts/egress.md) |
| Plugin model | D-11, D-16, D-42 to D-49, D-59 | Manifests, grants, effective permissions and workload compilation in `internal/plugins`. The connector contract (`openfief.connector.v1`) and the Go SDK for plugin authors in `sdk/go` | [Plugins](concepts/plugins.md) |
| Ledger | D-18, D-19, D-23, D-33 to D-35 | Domain, use cases and the PostgreSQL store in `internal/ledger`, including the guard against seeding an account twice; the core service runs it | [Ledger](concepts/ledger.md) |
| Valuation and FX | D-20, D-21, D-25, D-36 | Prices, FX lookups, valuation bases and the NBP client in `internal/valuation`; stored valuations and the River revaluation queue (D-36) in `internal/valuation/postgres` | [Valuation](concepts/valuation.md) |
| Treasury bonds | D-9, D-25, D-37 | Polish bond rules in `internal/jurisdiction/pl` | [Treasury bonds](concepts/treasury-bonds.md) |
| Ingest | D-7, D-8, D-23, D-49, D-59 | Import batches and the IBKR Flex mapping and pull in `internal/ingest`, still a package of the core rather than a plugin. The core's client for connector plugins is in `internal/ingest/connectrpc`, but the per-instance tokens it needs (D-59) aren't built | [Ingest](concepts/ingest.md) |
| Shared kernel | (code organization) | Money, currency, quantity and dates in `internal/kernel` | [Kernel](concepts/kernel.md) |
| Portfolio | D-52 to D-54, D-61 | Sleeves and sleeve kinds with their default roles and install defaults in `internal/portfolio`; the model portfolio is not built | [Portfolio](concepts/portfolio.md) |
| Demo dataset | GR-12 | A synthetic dataset in `internal/demo`, loaded by `cmd/demo-seed` through the same ingest, ledger and valuation paths real data takes. It refuses any database without its own demo marker | [Demo](dev/demo.md) |
| Read API | D-6, D-7, D-30, D-36, D-50, D-62 | The `openfief.views.v1.ReadService` protobuf service and its generated OpenAPI document; the queries, totals and allocation in `internal/views`; the Connect and REST `/v1` handler, through Vanguard, in `internal/views/connectapi`. The core service serves it | [Read API](concepts/read-api.md) |
| Core service | D-2, D-14, D-27, D-28, D-36, D-50 | `cmd/core`, installed by `deploy/core` (below). It migrates every context's schema, runs the River revaluation workers, and serves the read API on port 8080 with [script tokens](concepts/identity.md) from `internal/identity` as its only accepted tokens, plus unauthenticated `/healthz` and `/readyz`. `core token issue`, `list` and `revoke` manage the tokens. It doesn't call connectors yet (D-49): no connector plugin exists, so data enters through `cmd/demo-seed` | [Identity](concepts/identity.md), [Read API](concepts/read-api.md) |
| UI, Ghostfolio view, CLI, XTB parser, market data | D-3 to D-5, D-30, D-46, D-51 | Not built | [Status](/status/) |

The [status page](/status/) gives the same picture decision by decision, from the requirements in kibi.

## Bounded contexts

The code is organized by domain, following domain-driven design (Eric Evans, 2003). Each bounded context owns its terms, types and data. In this designed map from [code-organization.md](dev/code-organization.md), arrows point from a context to the contexts it reads:

```mermaid
flowchart LR
  ingest --> ledger
  ingest --> valuation
  valuation --> ledger
  valuation --> pl["jurisdiction/pl"]
  views --> ledger
  views --> valuation
  views --> portfolio
  plugins
  identity
  kernel["kernel (shared)"]
```

The [code map](/reference/code-map/) draws the same map from the actual imports on every build. Comparing the two is a quick architecture check. Today the code has two arrows the design lacks: `ledger` imports `portfolio` to attach accounts to sleeves, and `views` imports `jurisdiction/pl` to tell a treasury bond series from a listed instrument in the allocation.

## In the cluster

Everything lives in the `openfief` namespace. A cluster admin installs the custom resource definitions once. The operator's permissions are a Role in that namespace only (D-15). The install ships no Ingress, LoadBalancer or tunnel (D-28); only the demo environment adds its own HTTPS route, below.

```mermaid
flowchart LR
  subgraph cluster["Cluster"]
    crds["CRDs<br/>plugins.openfief.io<br/>plugingrants.openfief.io"]
    subgraph ns["namespace openfief"]
      op["Deployment openfief-operator<br/>Role: namespace only"]
      core["Deployment openfief-core<br/>Service openfief-core:<br/>http 8080, connect 8081"]
      db[("PostgreSQL<br/>CloudNativePG openfief-db")]
      px["Deployment egress-proxy<br/>Service egress-proxy"]
      p1["Plugin pod<br/>own ServiceAccount, Secrets,<br/>NetworkPolicy"]
      cm["ConfigMap<br/>egress allow-lists"]
    end
  end
  internet(["Internet"])
  core -- "DSN from Secret<br/>openfief-db-app" --> db
  p1 -- "connect port" --> core
  op -- "reads selector and ports" --> core
  op -- "watches" --> crds
  op -- "creates" --> p1
  op -- "writes" --> cm
  px -- "reads" --> cm
  p1 -- "HTTPS_PROXY, egress token<br/>as the proxy password" --> px
  px -- "granted hosts only" --> internet
```

The install manifests are in `deploy/`: `crds/` for the definitions and the namespace, `operator/`, `egress-proxy/` and `core/` for the three workloads. The core's Service `openfief-core` is a ClusterIP with two named ports on the same container port 8080: `http` (8080), the read API and health endpoints, reached with `kubectl port-forward` by default (D-28), and `connect` (8081), the port the operator hands to plugins (D-49). Its NetworkPolicy admits plugin pods and lets the core reach only DNS and the database; its pods mount no Kubernetes API token. The database URL comes from CloudNativePG's Secret `openfief-db-app`, key `uri` (D-14, D-29). The [operator](concepts/operator.md) and [egress](concepts/egress.md) pages walk through them.

## The demo environment

`deploy/demo` runs the same install on one Hetzner machine with k3s, holding synthetic data only (GR-12). Nothing pushes to it: cloud-init installs Flux, and Flux pulls each layer below, in dependency order. Five layers follow `main` in git. The `openfief` layer instead follows a signed OCI artifact that CI publishes after each push to `main`, with the images pinned to the digests whose signatures verified (D-17). The HTTPS route is the demo's own, on k3s's bundled Traefik with a Let's Encrypt certificate, to Service `openfief-core`, port `http`; a demo-only NetworkPolicy lets Traefik reach the core. Creating and destroying demos is in [demo.md](dev/demo.md).

```mermaid
flowchart LR
  git[("git: main")]
  oci[("signed OCI artifact<br/>digest-pinned images")]
  subgraph machine["Hetzner machine, k3s"]
    flux["Flux<br/>installed by cloud-init"]
    crds["demo-crds<br/>deploy/crds"]
    cnpg["demo-cnpg<br/>CloudNativePG operator"]
    db["demo-database<br/>one-instance cluster"]
    cm["demo-cert-manager"]
    ing["demo-ingress<br/>Let's Encrypt issuer"]
    routes["demo-routes<br/>Traefik route to the core"]
    of["demo-openfief<br/>operator, egress proxy,<br/>core, demo-seed Job"]
  end
  git --> flux
  oci --> flux
  flux --> crds & cnpg & cm
  crds --> db
  cnpg --> db
  cm --> ing --> routes
  crds --> of
  db --> of
  of --> routes
```

Arrows between layers are Flux `dependsOn` edges. The CI side, from signed images to the published artifact, is on the [delivery pipeline](dev/pipeline.md) page.

The synthetic data comes from `deploy/demo-data`, a Job that runs `cmd/demo-seed` against the demo database; `demo-openfief` applies it next to the core. The demo runs the Job without `--reset-demo` and the core without its default sleeves, so the two can start together; [demo.md](dev/demo.md#demo-data) explains why and how to refresh the data.

## Main flows

Each flow is drawn on the page that explains it:

- **Importing a broker statement:** the IBKR Flex pull or a connector plugin's delivery, mapping, and one ledger transaction per import. See [Ingest](concepts/ingest.md).
- **Revaluing after an import:** the import queues revaluation jobs in the same transaction, and a River worker stores the new valuations. See [Valuation](concepts/valuation.md).
- **Recording and deriving holdings:** versioned transactions, FIFO lots and balance assertions. See [Ledger](concepts/ledger.md).
- **Valuing a bond lot:** interest periods, indexation, early-redemption fee and Belka tax. See [Treasury bonds](concepts/treasury-bonds.md).
- **Running a plugin:** from a `Plugin` and its `PluginGrant` to an isolated pod. See [Operator](concepts/operator.md).
- **A plugin calling out:** token, allow-list and address checks in the proxy. See [Egress](concepts/egress.md).
- **From commit to release:** hooks, CI, signed images and the demo deploy. See [Delivery pipeline](dev/pipeline.md).
