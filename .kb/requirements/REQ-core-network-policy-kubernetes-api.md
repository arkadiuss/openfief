---
title: The core's NetworkPolicy must allow the Kubernetes API on port 6443 at any address.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-92-1
  - lane:strict
text_ref: design-decisions:D-92#1
origin:
  kind: import
  ref: design-decisions:D-92#1
  recorded_at: '2026-10-29T08:37:51.145Z'
semantic_text: The core's NetworkPolicy must allow the Kubernetes API on port 6443 at any address.
logic_claims:
  - CLAIM-80DB53D5E461D669
semantic_clauses:
  - The core's NetworkPolicy must allow the Kubernetes API on port 6443 at any address.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 532aed09690fa4267099c30070bffb3de229ebf6233329e4fabfb38421e6cf7a
semantic_inventory:
  - claim_key: CLAIM-80DB53D5E461D669
    claim_text: The core's NetworkPolicy must allow the Kubernetes API on port 6443 at any address
    role: normative
    status: modeled
    span:
      start: 0
      end: 82
id: REQ-core-network-policy-kubernetes-api
type: req
---
The core's NetworkPolicy must allow the Kubernetes API on port 6443 at any address.

## Context

The core calls the Kubernetes API, for TokenReviews (D-82, REQ-core-token-review-cluster-role) and for the Plugin resources it reads. Distributions such as k3s serve that API on the node's own address, which differs per cluster, so the author has the core's NetworkPolicy allow port 6443 at any address rather than a fixed one. REQ-docs-core-kubernetes-api-rule records the risk this leaves and how an owner narrows it. The author recorded D-91 and D-92 after the design session that answered the implementation's questions.

## Source

> The core's NetworkPolicy allows the Kubernetes API on port 6443 at any address, because distributions such as k3s serve it on the node's own address, which differs per cluster. The docs explain the risk, that a compromised core could use the port to reach other hosts, and how an owner narrows the rule to their API server's address.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-92#1
