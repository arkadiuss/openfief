---
title: On both kind and k3d, the operator must deploy a first-party plugin image built from the checkout and signed by an owner's trusted key, report the signature in the Plugin's status, run it hardened and reachable from the core only, publish exactly its granted egress hosts, and remove its workload and allow-list entry when the Plugin is deleted.
status: open
text_ref: docs/DESIGN.md#D-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-12
  - lane:strict
semantic_text: On both kind and k3d, the operator must deploy a first-party plugin image built from the checkout and signed by an owner's trusted key, report the signature in the Plugin's status, run it hardened and reachable from the core only, publish exactly its granted egress hosts, and remove its workload and allow-list entry when the Plugin is deleted.
logic_claims:
  - CLAIM-8B3035E46396EF6D
semantic_clauses:
  - On both kind and k3d, the operator must deploy a first-party plugin image built from the checkout and signed by an owner's trusted key, report the signature in the Plugin's status, run it hardened and reachable from the core only, publish exactly its granted egress hosts, and remove its workload and allow-list entry when the Plugin is deleted.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 70bb295db36de1c90afa69043d43ddb8d78fbfbaf547ce7933b2352ae17704de
semantic_inventory:
  - claim_key: CLAIM-8B3035E46396EF6D
    claim_text: On both kind and k3d, the operator must deploy a first-party plugin image built from the checkout and signed by an owner's trusted key, report the signature in the Plugin's status, run it hardened and reachable from the core only, publish exactly its granted egress hosts, and remove its workload and allow-list entry when the Plugin is deleted
    role: normative
    status: modeled
    span:
      start: 0
      end: 344
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:58.801Z'
id: REQ-e2e-operator-deploys-signed-plugin
type: req
---
On both kind and k3d, the operator must deploy a first-party plugin image built from the checkout and signed by an owner's trusted key, report the signature in the Plugin's status, run it hardened and reachable from the core only, publish exactly its granted egress hosts, and remove its workload and allow-list entry when the Plugin is deleted.

## Context

D-12 says OpenFief is tested on kind and k3d. The orchestrating agent asked to close the last D-12 gap: the end-to-end suite never deployed a plugin through the operator, because the operator refuses unsigned images (D-57). The suite now signs a plugin image with a key made for the run, trusted as an owner's signer (D-45), and checks the operator's verification (D-56), workload (D-27), NetworkPolicy (D-59) and allow-list (D-16, D-32). Modeled by Claude Code; the author has not reviewed it.

## Source

> **[D-12]** Any conformant Kubernetes cluster; no reference platform. Tested on kind and k3d, with documented requirements: a CNI that enforces NetworkPolicy, and a storage class.
>
> **[D-57]** An image without a valid signature from a trusted signer is never deployed, and there is no override. A developer testing a plugin locally signs it with their own key, added as a trusted signer (D-45).

docs/DESIGN.md#D-12
