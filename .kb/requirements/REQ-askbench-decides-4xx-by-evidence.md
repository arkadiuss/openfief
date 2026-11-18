---
title: 'The Ask bench decides whose fault every 4xx is by evidence and leaves none to nobody: a 400, 413, 422 or other refusal is the host''s fault when an identical request (the same messages, tools and response format, hashed without the model id, the provider routing, the reasoning control and the usage accounting) was accepted by another model@host entry of the same invocation, and the bench''s fault, which voids the run and records that the request check should have caught it, when no entry accepted it, each call recording the rule that decided it; a 404 is the host''s fault, counted against it and flagged in summary.md and stdout as a stale pin; a 401 or 402 voids the run as credentials or credit, counts against no host or model and makes the bench exit with its own code 5.'
status: open
semantic_text: 'The Ask bench decides whose fault every 4xx is by evidence and leaves none to nobody: a 400, 413, 422 or other refusal is the host''s fault when an identical request (the same messages, tools and response format, hashed without the model id, the provider routing, the reasoning control and the usage accounting) was accepted by another model@host entry of the same invocation, and the bench''s fault, which voids the run and records that the request check should have caught it, when no entry accepted it, each call recording the rule that decided it; a 404 is the host''s fault, counted against it and flagged in summary.md and stdout as a stale pin; a 401 or 402 voids the run as credentials or credit, counts against no host or model and makes the bench exit with its own code 5.'
firmness: default
required_evidence: unverified
tags:
  - askbench
origin:
  kind: agent
  ref: 'Brief of the Ask bench 4xx and per-question reliability change: the design delegate''s decisions on D-207 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T20:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7800a1e1e248f8592d69b08155cde336a7ec3f382149b93a5e5007ab3a652728
semantic_inventory:
  - claim_key: CLAIM-1A931344F9A82429
    claim_text: 'The Ask bench decides whose fault every 4xx is by evidence and leaves none to nobody: a 400, 413, 422 or other refusal is the host''s fault when an identical request (the same messages, tools and response format, hashed without the model id, the provider routing, the reasoning control and the usage accounting) was accepted by another model@host entry of the same invocation, and the bench''s fault, which voids the run and records that the request check should have caught it, when no entry accepted it, each call recording the rule that decided it'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 548
  - claim_key: CLAIM-BB9AB4D21FFC6485
    claim_text: a 404 is the host's fault, counted against it and flagged in summary.md and stdout as a stale pin
    role: descriptive
    status: missing
    span:
      start: 550
      end: 647
  - claim_key: CLAIM-70D373D36FA7E1A1
    claim_text: a 401 or 402 voids the run as credentials or credit, counts against no host or model and makes the bench exit with its own code 5
    role: descriptive
    status: missing
    span:
      start: 649
      end: 778
logic_claims:
  - CLAIM-1A931344F9A82429
  - CLAIM-BB9AB4D21FFC6485
  - CLAIM-70D373D36FA7E1A1
id: REQ-askbench-decides-4xx-by-evidence
type: req
---
The Ask bench decides whose fault every 4xx is by evidence and leaves none to nobody: a 400, 413, 422 or other refusal is the host's fault when an identical request (the same messages, tools and response format, hashed without the model id, the provider routing, the reasoning control and the usage accounting) was accepted by another model@host entry of the same invocation, and the bench's fault, which voids the run and records that the request check should have caught it, when no entry accepted it, each call recording the rule that decided it; a 404 is the host's fault, counted against it and flagged in summary.md and stdout as a stale pin; a 401 or 402 voids the run as credentials or credit, counts against no host or model and makes the bench exit with its own code 5.

## Context

After run 16-1 every failed call had a class and a fault, but a refused key, no credits and every other 4xx were left as nobody's and silently unscored. The design delegate decided that no 4xx may be nobody's: the evidence of the run says whose it is. A 400, 413 or 422 that another host accepted identically is that host's inability to serve the request (D-207); one that nobody accepted is a request the bench should not have sent. A 404 is a stale route pin. A refused key or no credits says nothing of any host or model, so the run is void and the weekly workflow must fail (D-221), with an exit code of its own.

## Source

The design delegate's decisions on the Ask bench, 2026-11-15: "No 4xx may be neither"; "400, 413, 422: if an identical request was accepted by another host/model entry in the same bench invocation, it is a host fault; if no entry accepted it, it is a bench fault: void the run"; "404: host / route-config fault, flagged loudly: the pin may be stale"; "401, 402: void the run as void (credentials/credit), never counts against any host or model; the bench exits non-zero with a distinct exit code so the weekly workflow fails and the author sees it (D-221)".