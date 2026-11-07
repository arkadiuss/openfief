---
title: openfief plugin ca-bundle set <instance> <pem-file> must accept only a file of PEM certificates, then create or rotate the Secret <instance>--ca-bundle (key ca.crt, labelled as the instance) and stamp the SHA-256 of what it holds on the Plugin spec, and with --print must emit both for GitOps with a note that the owner keeps the checksum in step with the Secret.
status: open
text_ref: design-session:connector-endpoint-ca-bundle-cli
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-bundle-cli
  - lane:strict
semantic_text: openfief plugin ca-bundle set <instance> <pem-file> must accept only a file of PEM certificates, then create or rotate the Secret <instance>--ca-bundle (key ca.crt, labelled as the instance) and stamp the SHA-256 of what it holds on the Plugin spec, and with --print must emit both for GitOps with a note that the owner keeps the checksum in step with the Secret.
logic_claims:
  - CLAIM-58C06F6FE963BCF7
semantic_clauses:
  - openfief plugin ca-bundle set <instance> <pem-file> must accept only a file of PEM certificates, then create or rotate the Secret <instance>--ca-bundle (key ca.crt, labelled as the instance) and stamp the SHA-256 of what it holds on the Plugin spec, and with --print must emit both for GitOps with a note that the owner keeps the checksum in step with the Secret.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 61a3a2946743c2a87d837d10816499baa33292d25a3116ccb95068d4371dce40
semantic_inventory:
  - claim_key: CLAIM-58C06F6FE963BCF7
    claim_text: openfief plugin ca-bundle set <instance> <pem-file> must accept only a file of PEM certificates, then create or rotate the Secret <instance>--ca-bundle (key ca.crt, labelled as the instance) and stamp the SHA-256 of what it holds on the Plugin spec, and with --print must emit both for GitOps with a note that the owner keeps the checksum in step with the Secret
    role: normative
    status: modeled
    span:
      start: 0
      end: 362
origin:
  kind: agent
  recorded_at: '2026-11-07T17:33:04.913Z'
id: REQ-plugins-cli-ca-bundle-set
type: req
---
openfief plugin ca-bundle set <instance> <pem-file> must accept only a file of PEM certificates, then create or rotate the Secret <instance>--ca-bundle (key ca.crt, labelled as the instance) and stamp the SHA-256 of what it holds on the Plugin spec, and with --print must emit both for GitOps with a note that the owner keeps the checksum in step with the Secret.

## Context

Design session refinement of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots, expected only when the instance setting extraCa says so), given to the implementation agent by the orchestrator on 2026-11-07 as option (a): the operator keeps no Secret access at all (D-15, D-29), so the owner CLI writes the bundle Secret and stamps its SHA-256 on the Plugin spec, and the operator reads only the spec. It replaces an earlier build in which the owner install let the operator get that one Secret through a per-instance Role and re-read it every minute. The owner CLI writes with the owner kubeconfig (D-46); the label lets uninstall delete it (D-182). Validation parses every PEM block and refuses anything else, an empty file, or non-CERTIFICATE blocks.

## Source

> openfief plugin ca-bundle set <instance> <pem-file>: validates the file is PEM certificates; creates or rotates <instance>--ca-bundle (key ca.crt) with the openfief.io/instance label; stamps the checksum on the Plugin spec; --print emits both the Secret and the Plugin change for GitOps, with a note that the owner must keep the checksum in step with the Secret.

design-session:connector-endpoint-ca-bundle-cli
