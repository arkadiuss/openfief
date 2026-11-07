---
title: openfief plugin ca-bundle unset <instance> must, once the owner confirms, delete the Secret <instance>--ca-bundle when it is labelled as the instance and clear the checksum stamped on its Plugin.
status: open
text_ref: design-session:connector-endpoint-ca-bundle-cli
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-bundle-cli
  - lane:strict
semantic_text: openfief plugin ca-bundle unset <instance> must, once the owner confirms, delete the Secret <instance>--ca-bundle when it is labelled as the instance and clear the checksum stamped on its Plugin.
logic_claims:
  - CLAIM-919655F0EDB0E67A
semantic_clauses:
  - openfief plugin ca-bundle unset <instance> must, once the owner confirms, delete the Secret <instance>--ca-bundle when it is labelled as the instance and clear the checksum stamped on its Plugin.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a313dd0f77a4a876945a01c1fd2076e42389f9ea25278bee94f7be9a354e8199
semantic_inventory:
  - claim_key: CLAIM-919655F0EDB0E67A
    claim_text: openfief plugin ca-bundle unset <instance> must, once the owner confirms, delete the Secret <instance>--ca-bundle when it is labelled as the instance and clear the checksum stamped on its Plugin
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
origin:
  kind: agent
  recorded_at: '2026-11-07T17:33:07.541Z'
id: REQ-plugins-cli-ca-bundle-unset
type: req
---
openfief plugin ca-bundle unset <instance> must, once the owner confirms, delete the Secret <instance>--ca-bundle when it is labelled as the instance and clear the checksum stamped on its Plugin.

## Context

Design session refinement of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots, expected only when the instance setting extraCa says so), given to the implementation agent by the orchestrator on 2026-11-07 as option (a): the operator keeps no Secret access at all (D-15, D-29), so the owner CLI writes the bundle Secret and stamps its SHA-256 on the Plugin spec, and the operator reads only the spec. It replaces an earlier build in which the owner install let the operator get that one Secret through a per-instance Role and re-read it every minute. Like uninstall, it deletes only a Secret labelled as the instance (D-182).

## Source

> openfief plugin ca-bundle unset <instance>: deletes the Secret and clears the stamp.

design-session:connector-endpoint-ca-bundle-cli
