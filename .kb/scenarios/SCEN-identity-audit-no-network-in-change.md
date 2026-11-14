---
title: A change asks the connector before its transaction and tells SpiceDB after it, and a network call inside a change is refused
status: active
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:03.726Z'
id: SCEN-identity-audit-no-network-in-change
type: scenario
---
Given an owner's change that needs a connector, SpiceDB or the cluster, when it runs, then the connector call and the manifest read are made before its transaction opens and its transaction holds only the writes, what SpiceDB is told is told after the commit, never when the change rolls back, and a network call made with the context of a change's transaction fails with ErrNetworkInChange before anything is sent.

## Context

The change's stores share one connection and hold their locks for as long as the change runs, so a slow call there would hold the database up. The owner's statement upload and bond lot entry and deletion ask the connector first (PrepareFile, PrepareBondLots) and only then run the import (Commit); what a change tells SpiceDB afterwards, such as an account's placement, is deferred until the outermost change commits. Every client the core builds over the network refuses a call made with a change's context, and an architecture test keeps the PostgreSQL adapters from importing a network client.

## Source

> NO NETWORK I/O INSIDE A CHANGE'S TRANSACTION (connector, model provider, SpiceDB, any Connect/HTTP client). Audit every outbox.Within / audited change path; move any such call before Begin and pass results in.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18