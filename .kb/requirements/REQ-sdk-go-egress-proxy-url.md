---
title: The Go SDK's egress client must send every https request through the egress proxy in HTTPS_PROXY, building the proxy URL on each request with the current token from OPENFIEF_EGRESS_TOKEN_FILE (audience openfief-egress) as its password, and must refuse plain http requests, requests without a proxy, and creating a client without an egress token file.
status: open
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
semantic_text: The Go SDK's egress client must send every https request through the egress proxy in HTTPS_PROXY, building the proxy URL on each request with the current token from OPENFIEF_EGRESS_TOKEN_FILE (audience openfief-egress) as its password, and must refuse plain http requests, requests without a proxy, and creating a client without an egress token file.
logic_claims:
  - CLAIM-D2E85C5DC0970981
semantic_clauses:
  - The Go SDK's egress client must send every https request through the egress proxy in HTTPS_PROXY, building the proxy URL on each request with the current token from OPENFIEF_EGRESS_TOKEN_FILE (audience openfief-egress) as its password, and must refuse plain http requests, requests without a proxy, and creating a client without an egress token file.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ff7bcc537b9737ab4f1e9a39dbb98b62a22d1d5cc90d216b6d7bf39d2821f63c
semantic_inventory:
  - claim_key: CLAIM-D2E85C5DC0970981
    claim_text: The Go SDK's egress client must send every https request through the egress proxy in HTTPS_PROXY, building the proxy URL on each request with the current token from OPENFIEF_EGRESS_TOKEN_FILE (audience openfief-egress) as its password, and must refuse plain http requests, requests without a proxy, and creating a client without an egress token file
    role: normative
    status: modeled
    span:
      start: 0
      end: 349
origin:
  kind: agent
  recorded_at: '2026-10-28T08:02:03.003Z'
id: REQ-sdk-go-egress-proxy-url
type: req
---
The Go SDK's egress client must send every https request through the egress proxy in HTTPS_PROXY, building the proxy URL on each request with the current token from OPENFIEF_EGRESS_TOKEN_FILE (audience openfief-egress) as its password, and must refuse plain http requests, requests without a proxy, and creating a client without an egress token file.

## Context

D-63 decides that a plugin sends its projected token as the password in the proxy URL, which Go sends as Proxy-Authorization, with audience openfief-egress, rebuilt from the token file on each request because tokens rotate. The variable OPENFIEF_EGRESS_TOKEN_FILE is the orchestrator's choice; the operator mounting that token is a separate task. Refines REQ-plugins-sdk-token-per-request and REQ-AUTO-8CBFBD6A12E0865B. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header [...] Projected tokens rotate, so the SDK builds the proxy URL (D-32's `HTTPS_PROXY`) from the token file on each request.

docs/DESIGN.md D-63
