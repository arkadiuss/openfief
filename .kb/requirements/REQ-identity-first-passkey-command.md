---
title: The openfief CLI must have a command for the first passkey that runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-80-1
  - lane:strict
text_ref: design-decisions:D-80#1
origin:
  kind: import
  ref: design-decisions:D-80#1
  recorded_at: '2026-10-29T08:09:53.496Z'
semantic_text: The openfief CLI must have a command for the first passkey that runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig.
logic_claims:
  - CLAIM-50D4CBD2D093CBF7
semantic_clauses:
  - The openfief CLI must have a command for the first passkey that runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: a6301616f8245b704c5c599096eeab6e2128dadaff74534bd46d664837033774
semantic_inventory:
  - claim_key: CLAIM-50D4CBD2D093CBF7
    claim_text: The openfief CLI must have a command for the first passkey that runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig
    role: normative
    status: modeled
    span:
      start: 0
      end: 169
id: REQ-identity-first-passkey-command
type: req
---
The openfief CLI must have a command for the first passkey that runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig.

## Context

D-80 answers the implementation's question about the first passkey from the openfief CLI: D-60 has the CLI print the one-time link for the first passkey (REQ-identity-install-without-hostname), but the core's enroll-link command prints it, runs inside the core's pod and needs the database, so the owner had to run it through kubectl exec. The one-time code it issues is REQ-identity-enrollment-code's. Like plugin installation (D-46), the command works with the owner's own kubeconfig. REQ-identity-first-passkey-link records the rest of D-80. The author recorded D-79 and D-80 after the design session that answered the implementation's questions.

## Source

> The `openfief` CLI has a command for the first passkey: it runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig, starts the `kubectl port-forward` to the UI, and prints D-60's one-time link. The first-run wizard (D-73) follows.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-80#1
