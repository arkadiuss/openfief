---
title: A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with --extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.
status: open
text_ref: design-session:connector-endpoint-ca-refinements
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-refinements
  - lane:strict
semantic_text: A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with --extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.
logic_claims:
  - CLAIM-2B3DBDFC3CA840A0
semantic_clauses:
  - A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with --extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ac152d408a8a555541e7b2bdcb251001ddabccdecda674e9f9e45b8ce1410ec6
semantic_inventory:
  - claim_key: CLAIM-2B3DBDFC3CA840A0
    claim_text: A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with --extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves
    role: normative
    status: modeled
    span:
      start: 0
      end: 281
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:05.889Z'
id: REQ-plugins-extra-ca-setting
type: req
---
A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with --extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.

## Context

Design session refinements of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots), given to the implementation agent by the orchestrator on 2026-11-07, with an update that the expectation of a bundle must be explicit, a boolean extraCA instance setting, never inferred from a base URL that differs from its default, since mirrors and other providers may have public certificates. The setting lives on the Plugin spec next to the in-cluster service (D-74), not in the plugin config, so no plugin config schema (additionalProperties false) has to repeat it. Builds on D-89 (instance Secrets by role) and D-16 (the grant approves the Secret either way).

## Source

> Make the intent explicit: a boolean in the instance config, e.g. extraCA: true, saying a bundle is expected. Mount the ca-bundle Secret and stamp its checksum on the pod template only when extraCA is set.

design-session:connector-endpoint-ca-refinements
