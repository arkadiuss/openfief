---
title: When a running plugin instance expects an extra CA bundle and its ca-bundle Secret does not exist, the operator must set the Plugin condition CABundleMissing to True with what to do, False when the Secret exists, Unknown when it may not read it, and no condition when none is expected, never holding the Secret data and never keeping the plugin from running.
status: open
text_ref: design-session:connector-endpoint-ca-refinements
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-refinements
  - lane:strict
semantic_text: When a running plugin instance expects an extra CA bundle and its ca-bundle Secret does not exist, the operator must set the Plugin condition CABundleMissing to True with what to do, False when the Secret exists, Unknown when it may not read it, and no condition when none is expected, never holding the Secret data and never keeping the plugin from running.
logic_claims:
  - CLAIM-7C1D90B7861616A1
semantic_clauses:
  - When a running plugin instance expects an extra CA bundle and its ca-bundle Secret does not exist, the operator must set the Plugin condition CABundleMissing to True with what to do, False when the Secret exists, Unknown when it may not read it, and no condition when none is expected, never holding the Secret data and never keeping the plugin from running.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 63e8d6c8f57b6a836232ce87cf72fabcb54c1f542c6d8b68f6dd69563fcbc0c0
semantic_inventory:
  - claim_key: CLAIM-7C1D90B7861616A1
    claim_text: When a running plugin instance expects an extra CA bundle and its ca-bundle Secret does not exist, the operator must set the Plugin condition CABundleMissing to True with what to do, False when the Secret exists, Unknown when it may not read it, and no condition when none is expected, never holding the Secret data and never keeping the plugin from running
    role: condition
    status: modeled
    span:
      start: 0
      end: 357
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:10.246Z'
id: REQ-operator-ca-bundle-missing-condition
type: req
---
When a running plugin instance expects an extra CA bundle and its ca-bundle Secret does not exist, the operator must set the Plugin condition CABundleMissing to True with what to do, False when the Secret exists, Unknown when it may not read it, and no condition when none is expected, never holding the Secret data and never keeping the plugin from running.

## Context

Design session refinements of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots), given to the implementation agent by the orchestrator on 2026-11-07, with an update that the expectation of a bundle must be explicit, a boolean extraCA instance setting, never inferred from a base URL that differs from its default, since mirrors and other providers may have public certificates. So the owner sees why TLS to a custom endpoint fails instead of it failing silently. GR-12 keeps Secret contents out of status, logs and events; the bundle stays optional, nothing waits for it.

## Source

> Raise CABundleMissing only when extraCA is set and <instance>--ca-bundle is absent. Otherwise behaviour is unchanged; still informational, never blocks deployment.

design-session:connector-endpoint-ca-refinements
