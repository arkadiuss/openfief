---
title: Local development of the UI must run against a local cluster such as kind or k3d.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-84-2
  - lane:strict
text_ref: design-decisions:D-84#2
origin:
  kind: import
  ref: design-decisions:D-84#2
  recorded_at: '2026-10-29T08:22:10.211Z'
semantic_text: Local development of the UI must run against a local cluster such as kind or k3d.
logic_claims:
  - CLAIM-A19F765882B2B26D
semantic_clauses:
  - Local development of the UI must run against a local cluster such as kind or k3d.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: d1a80de941afc235a006522c9f8fce54c86e419871ae480db6417002f284c750
semantic_inventory:
  - claim_key: CLAIM-A19F765882B2B26D
    claim_text: Local development of the UI must run against a local cluster such as kind or k3d
    role: normative
    status: modeled
    span:
      start: 0
      end: 80
id: REQ-ui-local-development-cluster
type: req
---
Local development of the UI must run against a local cluster such as kind or k3d.

## Context

D-84 forbids a fixed token or any other way around authentication in every build, even for development (REQ-identity-no-auth-bypass). The built-in UI calls the core's AuthService with its own projected ServiceAccount token (D-69, REQ-identity-auth-callers), which exists only in a cluster, so developing the UI happens against a local Kubernetes cluster such as kind or k3d rather than a UI process outside a cluster. The author recorded D-83 to D-85 after the design session that answered the implementation's questions.

## Source

> No build of OpenFief has a fixed token or any other way around authentication, not even for development. Local development of the UI runs against a local cluster such as kind or k3d.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-84#2
