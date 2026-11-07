---
title: A plugin manifest may mark a Secret role optional; the operator must mount that Secret as an optional volume so the plugin pod starts while the owner has not created it, every other Secret must stay required, the flag must survive the Plugin resource the CLI writes and reads, and the install summary must say the plugin runs without it.
status: open
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
semantic_text: A plugin manifest may mark a Secret role optional; the operator must mount that Secret as an optional volume so the plugin pod starts while the owner has not created it, every other Secret must stay required, the flag must survive the Plugin resource the CLI writes and reads, and the install summary must say the plugin runs without it.
logic_claims:
  - CLAIM-E2AAB6F33C41BABF
semantic_clauses:
  - A plugin manifest may mark a Secret role optional; the operator must mount that Secret as an optional volume so the plugin pod starts while the owner has not created it, every other Secret must stay required, the flag must survive the Plugin resource the CLI writes and reads, and the install summary must say the plugin runs without it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 512578b5bd89bb56ebdbaf4000b4e92ad3ef22d1ddb2a12c3be4cc6bc5b6f8e6
semantic_inventory:
  - claim_key: CLAIM-E2AAB6F33C41BABF
    claim_text: A plugin manifest may mark a Secret role optional; the operator must mount that Secret as an optional volume so the plugin pod starts while the owner has not created it, every other Secret must stay required, the flag must survive the Plugin resource the CLI writes and reads, and the install summary must say the plugin runs without it
    role: normative
    status: modeled
    span:
      start: 0
      end: 336
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:04.922Z'
id: REQ-plugins-optional-secret-mount
type: req
---
A plugin manifest may mark a Secret role optional; the operator must mount that Secret as an optional volume so the plugin pod starts while the owner has not created it, every other Secret must stay required, the flag must survive the Plugin resource the CLI writes and reads, and the install summary must say the plugin runs without it.

## Context

Design session decision for M0 exit end-to-end tests (D-200) and later model providers (D-205), given to the implementation agent by the orchestrator on 2026-11-06: connectors get two config fields, a schema-validated https base URL defaulting to the real host, and an optional extra CA bundle from a Secret added to the system roots, with no test-only switches. The CA bundle Secret is optional, so the manifest needs a way to say so (D-89 names the Secret <instance>--ca-bundle).

## Source

> extra CA bundle from a Secret (optional): PEM bundle ADDED to the system roots, never replacing them; mounted like other instance Secrets (role-named per D-89, e.g. <instance>--ca-bundle)

design-session:connector-endpoint-ca
