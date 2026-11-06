---
title: Each first-party network connector, ibkr-flex, eodhd, nbp-fx and treasury-bonds for both GUS and NBP, must take an optional https base URL config field that its config schema validates and that defaults to the provider's real host, must call that base URL instead of a hard-coded host, must name the field in egressFromConfig so the grant approves a configured host, and must request the optional ca-bundle Secret its endpoint client trusts besides the system roots.
status: open
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
semantic_text: Each first-party network connector, ibkr-flex, eodhd, nbp-fx and treasury-bonds for both GUS and NBP, must take an optional https base URL config field that its config schema validates and that defaults to the provider's real host, must call that base URL instead of a hard-coded host, must name the field in egressFromConfig so the grant approves a configured host, and must request the optional ca-bundle Secret its endpoint client trusts besides the system roots.
logic_claims:
  - CLAIM-403A535044B4CC17
semantic_clauses:
  - Each first-party network connector, ibkr-flex, eodhd, nbp-fx and treasury-bonds for both GUS and NBP, must take an optional https base URL config field that its config schema validates and that defaults to the provider's real host, must call that base URL instead of a hard-coded host, must name the field in egressFromConfig so the grant approves a configured host, and must request the optional ca-bundle Secret its endpoint client trusts besides the system roots.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a0dc2ad9fd21c6192ec81827538c1fa1960dbb4031a10d99e8dee82bc81cc32e
semantic_inventory:
  - claim_key: CLAIM-403A535044B4CC17
    claim_text: Each first-party network connector, ibkr-flex, eodhd, nbp-fx and treasury-bonds for both GUS and NBP, must take an optional https base URL config field that its config schema validates and that defaults to the provider's real host, must call that base URL instead of a hard-coded host, must name the field in egressFromConfig so the grant approves a configured host, and must request the optional ca-bundle Secret its endpoint client trusts besides the system roots
    role: normative
    status: modeled
    span:
      start: 0
      end: 465
origin:
  kind: agent
  recorded_at: '2026-11-06T22:47:55.033Z'
id: REQ-connectors-configurable-endpoint
type: req
---
Each first-party network connector, ibkr-flex, eodhd, nbp-fx and treasury-bonds for both GUS and NBP, must take an optional https base URL config field that its config schema validates and that defaults to the provider's real host, must call that base URL instead of a hard-coded host, must name the field in egressFromConfig so the grant approves a configured host, and must request the optional ca-bundle Secret its endpoint client trusts besides the system roots.

## Context

Design session decision for M0 exit end-to-end tests (D-200) and later model providers (D-205), given to the implementation agent by the orchestrator on 2026-11-06: connectors get two config fields, a schema-validated https base URL defaulting to the real host, and an optional extra CA bundle from a Secret added to the system roots, with no test-only switches. A configured host reaches the egress allow-list as a config host (D-75): the install reads it from the config, shows it, and the grant approves it beside the manifest fixed host.

## Source

> Apply it to the network connectors: ibkr-flex, eodhd, nbp-fx, treasury-bonds (and GUS/NBP calls inside treasury-bonds). The CA bundle Secret role in each manifest secret roles (optional).

design-session:connector-endpoint-ca
