---
title: When a plugin base URL moves to a host its grant does not approve, the running version must keep the old approved egress allow-list with the new host pending approval, and the upgrade diff must show the new host added and the old one removed.
status: open
text_ref: design-session:connector-endpoint-ca-refinements
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-refinements
  - lane:strict
semantic_text: When a plugin base URL moves to a host its grant does not approve, the running version must keep the old approved egress allow-list with the new host pending approval, and the upgrade diff must show the new host added and the old one removed.
logic_claims:
  - CLAIM-0924ABCEA06E85BE
semantic_clauses:
  - When a plugin base URL moves to a host its grant does not approve, the running version must keep the old approved egress allow-list with the new host pending approval, and the upgrade diff must show the new host added and the old one removed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 511982eaff79eedba5dd7dbd1fba97d540327a4c4fadd17bc87b5f9b99054acb
semantic_inventory:
  - claim_key: CLAIM-0924ABCEA06E85BE
    claim_text: When a plugin base URL moves to a host its grant does not approve, the running version must keep the old approved egress allow-list with the new host pending approval, and the upgrade diff must show the new host added and the old one removed
    role: condition
    status: modeled
    span:
      start: 0
      end: 241
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:19.130Z'
id: REQ-plugins-config-host-change-keeps-approved
type: req
---
When a plugin base URL moves to a host its grant does not approve, the running version must keep the old approved egress allow-list with the new host pending approval, and the upgrade diff must show the new host added and the old one removed.

## Context

Design session refinements of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots), given to the implementation agent by the orchestrator on 2026-11-07, with an update that the expectation of a bundle must be explicit, a boolean extraCA instance setting, never inferred from a base URL that differs from its default, since mirrors and other providers may have public certificates. Confirms D-16 (only what both Plugin and grant hold is enforced), D-75 (config hosts are approved like fixed ones) and D-90 (an upgrade shows how requested permissions differ) for a base URL change.

## Source

> Also confirm that while a new grant is pending after a base-URL host change, the operator keeps enforcing the OLD approved allow-list (D-16/D-90) and the install/upgrade diff shows the host change.

design-session:connector-endpoint-ca-refinements
