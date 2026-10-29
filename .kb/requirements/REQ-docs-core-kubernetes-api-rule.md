---
title: The docs must explain that a compromised core could use the Kubernetes API port rule to reach other hosts, and how an owner narrows the rule to their API server's address.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-92-2
  - lane:strict
text_ref: design-decisions:D-92#2
origin:
  kind: import
  ref: design-decisions:D-92#2
  recorded_at: '2026-10-29T08:38:32.553Z'
semantic_text: The docs must explain that a compromised core could use the Kubernetes API port rule to reach other hosts, and how an owner narrows the rule to their API server's address.
logic_claims:
  - CLAIM-3A2FEB02FB8C2B30
semantic_clauses:
  - The docs must explain that a compromised core could use the Kubernetes API port rule to reach other hosts, and how an owner narrows the rule to their API server's address.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 632a12daa452947b784d81da19f9cf9445165dc1d3dfc9bd973b715c3d6ecc24
semantic_inventory:
  - claim_key: CLAIM-3A2FEB02FB8C2B30
    claim_text: The docs must explain that a compromised core could use the Kubernetes API port rule to reach other hosts, and how an owner narrows the rule to their API server's address
    role: normative
    status: modeled
    span:
      start: 0
      end: 170
id: REQ-docs-core-kubernetes-api-rule
type: req
---
The docs must explain that a compromised core could use the Kubernetes API port rule to reach other hosts, and how an owner narrows the rule to their API server's address.

## Context

D-92 has the core's NetworkPolicy allow port 6443 at any address so the core reaches the Kubernetes API on any distribution (REQ-core-network-policy-kubernetes-api). That rule also lets a compromised core reach any other host listening on that port, so the author asks the docs to state the risk and show how an owner who knows their API server's address narrows the rule to it. The author recorded D-91 and D-92 after the design session that answered the implementation's questions.

## Source

> The core's NetworkPolicy allows the Kubernetes API on port 6443 at any address, because distributions such as k3s serve it on the node's own address, which differs per cluster. The docs explain the risk, that a compromised core could use the port to reach other hosts, and how an owner narrows the rule to their API server's address.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-92#2
