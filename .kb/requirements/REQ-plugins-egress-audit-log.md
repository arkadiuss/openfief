---
title: The egress audit must fail when the egress proxy allowed a request to a host the asking instance's grant does not approve, or when a decision line holds a field beyond who asked, where to, the decision and its timing, a requested host that is not a bare host and port, or no instance or decision; it must only read, and must report counts, instances and hosts but never a value that could hold a request's contents.
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: The egress audit must fail when the egress proxy allowed a request to a host the asking instance's grant does not approve, or when a decision line holds a field beyond who asked, where to, the decision and its timing, a requested host that is not a bare host and port, or no instance or decision; it must only read, and must report counts, instances and hosts but never a value that could hold a request's contents.
logic_claims:
  - CLAIM-21092876C272547E
semantic_clauses:
  - The egress audit must fail when the egress proxy allowed a request to a host the asking instance's grant does not approve, or when a decision line holds a field beyond who asked, where to, the decision and its timing, a requested host that is not a bare host and port, or no instance or decision; it must only read, and must report counts, instances and hosts but never a value that could hold a request's contents.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2008f5f8115bdcfff1f99ee5a9241da2366ea455259731ad3f41da3ce8eda5d1
semantic_inventory:
  - claim_key: CLAIM-21092876C272547E
    claim_text: The egress audit must fail when the egress proxy allowed a request to a host the asking instance's grant does not approve, or when a decision line holds a field beyond who asked, where to, the decision and its timing, a requested host that is not a bare host and port, or no instance or decision; it must only read, and must report counts, instances and hosts but never a value that could hold a request's contents
    role: normative
    status: modeled
    span:
      start: 0
      end: 414
origin:
  kind: agent
  recorded_at: '2026-11-07T03:31:02.627Z'
id: REQ-plugins-egress-audit-log
type: req
---
The egress audit must fail when the egress proxy allowed a request to a host the asking instance's grant does not approve, or when a decision line holds a field beyond who asked, where to, the decision and its timing, a requested host that is not a bare host and port, or no instance or decision; it must only read, and must report counts, instances and hosts but never a value that could hold a request's contents.

## Context

D-200's criterion 16 verifies destinations from the egress proxy's logs, and GR-12 keeps account data out of logs; the CI suite can't open an allowed tunnel (D-184), so the same audit runs on the demo and in the release smoke test, and the end-to-end suite uses it too. The orchestrator asked for it on 2026-11-07. Modeled by Claude Code for the D-200 follow-ups; the author has not reviewed the wording.

## Source

> **Egress.** Outbound traffic goes only to configured destinations (IBKR's Flex service, NBP, GUS, the market-data vendor, the user's Ghostfolio), verified from the egress proxy's logs.

design-decisions:D-200#16 (docs/DESIGN.md, D-200, criterion 16); GR-12
