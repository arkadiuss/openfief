---
title: The end-to-end suite's cluster configuration may exempt one host from the egress proxy's refused address ranges, and its NetworkPolicy lets the proxy reach that host.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-215-1
  - lane:strict
text_ref: design-decisions:D-215#1
origin:
  kind: import
  ref: design-decisions:D-215#1
  recorded_at: '2026-11-12T23:18:58.590Z'
semantic_text: The end-to-end suite's cluster configuration may exempt one host from the egress proxy's refused address ranges, and its NetworkPolicy lets the proxy reach that host.
logic_claims:
  - CLAIM-0FB258547E947F44
semantic_clauses:
  - The end-to-end suite's cluster configuration may exempt one host from the egress proxy's refused address ranges, and its NetworkPolicy lets the proxy reach that host.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 69732f427863fd258c3270a032163ea348e67fc6aff09006b14f8dd623ed305a
semantic_inventory:
  - claim_key: CLAIM-0FB258547E947F44
    claim_text: The end-to-end suite's cluster configuration may exempt one host from the egress proxy's refused address ranges, and its NetworkPolicy lets the proxy reach that host
    role: exception
    status: modeled
    span:
      start: 0
      end: 165
id: REQ-plugins-egress-test-exemption-one-host
type: req
---
The end-to-end suite's cluster configuration may exempt one host from the egress proxy's refused address ranges, and its NetworkPolicy lets the proxy reach that host.

## Context

D-215 is an exception to D-32 for tests alone, so that D-200's criterion 16 (REQ-m0-exit-16-egress) is checked on an allowed connection end to end: the proxy refuses private addresses (REQ-plugins-egress-forbidden-addresses), and a cluster's Services are private, so without an exemption no connection through the proxy could succeed in the suite. The exemption names exactly one host and the addresses it must resolve to; another host, the same host at another address and the host's address without its name all stay refused. The author recorded D-215 in DESIGN.md.

## Source

> An exception to D-32 for tests alone: the end-to-end suite's cluster configuration may exempt one host from the egress proxy's refused address ranges, and its NetworkPolicy lets the proxy reach that host, so D-200's criterion 16 is checked on an allowed connection end to end.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-215#1
