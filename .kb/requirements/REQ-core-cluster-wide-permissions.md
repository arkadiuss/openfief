---
title: OpenFief's only cluster-wide permissions must be the core's and the egress proxy's ClusterRoles.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-82-3
  - lane:strict
text_ref: design-decisions:D-82#3
origin:
  kind: import
  ref: design-decisions:D-82#3
  recorded_at: '2026-10-29T08:17:20.645Z'
semantic_text: OpenFief's only cluster-wide permissions must be the core's and the egress proxy's ClusterRoles.
logic_claims:
  - CLAIM-C9CB64215EFB73F0
semantic_clauses:
  - OpenFief's only cluster-wide permissions must be the core's and the egress proxy's ClusterRoles.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7dcf0550a7976b82822ad3a756647b40a4dadf1f051d70bb2e471cfd67594d57
semantic_inventory:
  - claim_key: CLAIM-C9CB64215EFB73F0
    claim_text: OpenFief's only cluster-wide permissions must be the core's and the egress proxy's ClusterRoles
    role: normative
    status: modeled
    span:
      start: 0
      end: 95
id: REQ-core-cluster-wide-permissions
type: req
---
OpenFief's only cluster-wide permissions must be the core's and the egress proxy's ClusterRoles.

## Context

D-82 gives the core a ClusterRole that allows only creating TokenReviews (REQ-core-token-review-cluster-role), and the egress proxy already checks plugins' egress tokens with TokenReview (D-63, REQ-plugins-egress-proxy-identifies-plugin). The author states that these two are the only cluster-wide permissions OpenFief holds; everything else, including the operator (D-15, REQ-AUTO-BBBDDCF207E884A1), stays inside OpenFief's namespace. The author recorded D-81 and D-82 after the design session that answered the implementation's questions.

## Source

> The core checks view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists. Its ClusterRole allows only creating TokenReviews; with the egress proxy's (D-63), it is one of the two cluster-wide permissions OpenFief holds.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-82#3
