---
title: 'The operator must set the Plugin condition CABundleMissing from the spec alone: True with reason ChecksumNotStamped when extraCa is set and no caBundleSha256 is stamped, False when one is, and no condition when extraCa is not set, never reading the Secret and never keeping the plugin from running.'
status: open
text_ref: design-session:connector-endpoint-ca-bundle-cli
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-bundle-cli
  - lane:strict
semantic_text: 'The operator must set the Plugin condition CABundleMissing from the spec alone: True with reason ChecksumNotStamped when extraCa is set and no caBundleSha256 is stamped, False when one is, and no condition when extraCa is not set, never reading the Secret and never keeping the plugin from running.'
logic_claims:
  - CLAIM-675338357AD22ADF
semantic_clauses:
  - 'The operator must set the Plugin condition CABundleMissing from the spec alone: True with reason ChecksumNotStamped when extraCa is set and no caBundleSha256 is stamped, False when one is, and no condition when extraCa is not set, never reading the Secret and never keeping the plugin from running.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: af410295576de1f9db313e5f69ce8f9dd680394354b30c55abf0d603b8ec70ff
semantic_inventory:
  - claim_key: CLAIM-675338357AD22ADF
    claim_text: 'The operator must set the Plugin condition CABundleMissing from the spec alone: True with reason ChecksumNotStamped when extraCa is set and no caBundleSha256 is stamped, False when one is, and no condition when extraCa is not set, never reading the Secret and never keeping the plugin from running'
    role: normative
    status: modeled
    span:
      start: 0
      end: 297
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:10.246Z'
id: REQ-operator-ca-bundle-missing-condition
type: req
---
The operator must set the Plugin condition CABundleMissing from the spec alone: True with reason ChecksumNotStamped when extraCa is set and no caBundleSha256 is stamped, False when one is, and no condition when extraCa is not set, never reading the Secret and never keeping the plugin from running.

## Context

Design session refinement of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots, expected only when the instance setting extraCa says so), given to the implementation agent by the orchestrator on 2026-11-07 as option (a): the operator keeps no Secret access at all (D-15, D-29), so the owner CLI writes the bundle Secret and stamps its SHA-256 on the Plugin spec, and the operator reads only the spec. It replaces an earlier build in which the owner install let the operator get that one Secret through a per-instance Role and re-read it every minute. So the owner sees that no bundle was set instead of TLS failing silently; a stamp that does not match the Secret shows up only as the plugin TLS failures, never as Secret reads (GR-12, D-15, D-29).

## Source

> Condition CABundleMissing read from the spec alone: True (reason e.g. ChecksumNotStamped) when extraCa is set and no checksum is stamped; False when stamped; absent when extraCa is not set. Never blocks deployment.

design-session:connector-endpoint-ca-bundle-cli
