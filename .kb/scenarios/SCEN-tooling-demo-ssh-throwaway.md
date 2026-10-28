---
title: Without a key, the server gets a throwaway key nobody holds
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:33.018Z'
id: SCEN-tooling-demo-ssh-throwaway
type: scenario
---
Given no SSH key labelled openfief-ssh=true and DEMO_SSH_CIDRS empty, or DEMO_SSH=off, when task demo:up creates the server, then an Ed25519 public key named openfief-demo-throwaway, labelled app=openfief, env=demo and openfief-ssh=throwaway, is uploaded and is the server's only key, the firewall has no SSH rule, and no private key is written. A later up reuses an existing throwaway key; down deletes it after the server. DEMO_SSH=off with SSH sources, DEMO_SSH=keys without a key and an unknown DEMO_SSH are refused before anything is created.