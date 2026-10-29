---
title: The core must check view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-82-1
  - lane:strict
text_ref: design-decisions:D-82#1
origin:
  kind: import
  ref: design-decisions:D-82#1
  recorded_at: '2026-10-29T08:15:29.566Z'
semantic_text: The core must check view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists.
logic_claims:
  - CLAIM-32EA50C847B50C55
semantic_clauses:
  - The core must check view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 73dc24cf8a696994ccbd3c2ed61d2259d9d6c70733934da1f5357af6644144f7
semantic_inventory:
  - claim_key: CLAIM-32EA50C847B50C55
    claim_text: The core must check view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists
    role: normative
    status: modeled
    span:
      start: 0
      end: 129
id: REQ-identity-view-plugin-token-review
type: req
---
The core must check view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists.

## Context

D-27 has plugins authenticate to the core with projected ServiceAccount tokens (REQ-AUTO-C6544ADAF8F5AE65), and D-69 offers the core's authentication service to view plugins only (REQ-identity-auth-callers). D-82 answers the implementation's question about the core's own cluster role for TokenReview: the core already checks view plugins' tokens that way (REQ-identity-read-api-plugins), and D-82 keeps it. REQ-core-token-review-cluster-role records the permission this needs. The author recorded D-81 and D-82 after the design session that answered the implementation's questions.

## Source

> The core checks view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists. Its ClusterRole allows only creating TokenReviews; with the egress proxy's (D-63), it is one of the two cluster-wide permissions OpenFief holds.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-82#1
