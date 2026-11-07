---
title: A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with install --extra-ca and toggles with upgrade --extra-ca or --no-extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.
status: open
text_ref: design-session:connector-endpoint-ca-bundle-cli
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-bundle-cli
  - lane:strict
semantic_text: A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with install --extra-ca and toggles with upgrade --extra-ca or --no-extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.
logic_claims:
  - CLAIM-0472AE5A88452CC9
semantic_clauses:
  - A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with install --extra-ca and toggles with upgrade --extra-ca or --no-extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c1535919b4ab81edffecf1dbdb285d670e65e2929fe3e5eaa9e6fb5b9004cd87
semantic_inventory:
  - claim_key: CLAIM-0472AE5A88452CC9
    claim_text: A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with install --extra-ca and toggles with upgrade --extra-ca or --no-extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves
    role: normative
    status: modeled
    span:
      start: 0
      end: 342
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:05.889Z'
id: REQ-plugins-extra-ca-setting
type: req
---
A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with install --extra-ca and toggles with upgrade --extra-ca or --no-extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.

## Context

Design session refinement of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots, expected only when the instance setting extraCa says so), given to the implementation agent by the orchestrator on 2026-11-07 as option (a): the operator keeps no Secret access at all (D-15, D-29), so the owner CLI writes the bundle Secret and stamps its SHA-256 on the Plugin spec, and the operator reads only the spec. It replaces an earlier build in which the owner install let the operator get that one Secret through a per-instance Role and re-read it every minute. The setting lives on the Plugin spec next to the in-cluster service (D-74), not in the plugin config. Builds on D-89 (instance Secrets by role), D-16 (the grant approves the Secret either way) and D-90 (upgrades). Turning it off clears the stamped checksum.

## Source

> openfief plugin upgrade --extra-ca / --no-extra-ca to toggle without reinstalling (refuse --extra-ca if the new manifest does not ask for the role).

design-session:connector-endpoint-ca-bundle-cli
