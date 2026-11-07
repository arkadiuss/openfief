---
title: While a plugin instance mounts its extra CA bundle, the operator must copy the checksum stamped on its Plugin spec (caBundleSha256) to the pod template annotation openfief.io/ca-bundle-sha256, so a new stamp rolls the pod, without reading the Secret or re-reading anything on a timer.
status: open
text_ref: design-session:connector-endpoint-ca-bundle-cli
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-bundle-cli
  - lane:strict
semantic_text: While a plugin instance mounts its extra CA bundle, the operator must copy the checksum stamped on its Plugin spec (caBundleSha256) to the pod template annotation openfief.io/ca-bundle-sha256, so a new stamp rolls the pod, without reading the Secret or re-reading anything on a timer.
logic_claims:
  - CLAIM-296F94DAD8F9D28A
semantic_clauses:
  - While a plugin instance mounts its extra CA bundle, the operator must copy the checksum stamped on its Plugin spec (caBundleSha256) to the pod template annotation openfief.io/ca-bundle-sha256, so a new stamp rolls the pod, without reading the Secret or re-reading anything on a timer.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9e636991967143a1de61b64f4ffbddf47e296e9224437f8c17d1ad7e8c9f12be
semantic_inventory:
  - claim_key: CLAIM-296F94DAD8F9D28A
    claim_text: While a plugin instance mounts its extra CA bundle, the operator must copy the checksum stamped on its Plugin spec (caBundleSha256) to the pod template annotation openfief.io/ca-bundle-sha256, so a new stamp rolls the pod, without reading the Secret or re-reading anything on a timer
    role: normative
    status: modeled
    span:
      start: 0
      end: 283
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:14.655Z'
id: REQ-operator-ca-bundle-rolls-pod
type: req
---
While a plugin instance mounts its extra CA bundle, the operator must copy the checksum stamped on its Plugin spec (caBundleSha256) to the pod template annotation openfief.io/ca-bundle-sha256, so a new stamp rolls the pod, without reading the Secret or re-reading anything on a timer.

## Context

Design session refinement of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots, expected only when the instance setting extraCa says so), given to the implementation agent by the orchestrator on 2026-11-07 as option (a): the operator keeps no Secret access at all (D-15, D-29), so the owner CLI writes the bundle Secret and stamps its SHA-256 on the Plugin spec, and the operator reads only the spec. It replaces an earlier build in which the owner install let the operator get that one Secret through a per-instance Role and re-read it every minute. The operator own Role reads no Secret (D-15, D-29), with no exception for the bundle; the CLI computes the checksum.

## Source

> The operator copies it to the pod template annotation openfief.io/ca-bundle-sha256, so a rotation rolls the pod.

design-session:connector-endpoint-ca-bundle-cli
