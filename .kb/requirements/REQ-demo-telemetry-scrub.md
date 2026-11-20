---
title: 'The demo''s telemetry collector scrubs visitors'' personal data before anything leaves it: the processor transform/scrub, which fails closed (error_mode propagate), runs in every logs, traces and metrics pipeline with nothing but the memory limiter before it; it drops the attributes that carry a client address, a session or token ID, a credential, Authorization, Cookie or another header, the user agent, or a question or prompt, and rewrites any IPv4 or IPv6 address, any openfief token prefix (ofs_, ofsess_, ofenroll_), Bearer or Basic credential, JSON web token and long token-like string, in the resource, scope, span, span event, log record and metric data point attributes, the log body and the span''s name and status message; scripts/demo/validate.sh checks the built layer and a unit test runs each statement on sample records.'
status: open
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data (GR-12, D-172, D-203, D-227), 2026-11-19'
  recorded_at: '2026-11-19T16:20:41.336Z'
semantic_text: 'The demo''s telemetry collector scrubs visitors'' personal data before anything leaves it: the processor transform/scrub, which fails closed (error_mode propagate), runs in every logs, traces and metrics pipeline with nothing but the memory limiter before it; it drops the attributes that carry a client address, a session or token ID, a credential, Authorization, Cookie or another header, the user agent, or a question or prompt, and rewrites any IPv4 or IPv6 address, any openfief token prefix (ofs_, ofsess_, ofenroll_), Bearer or Basic credential, JSON web token and long token-like string, in the resource, scope, span, span event, log record and metric data point attributes, the log body and the span''s name and status message; scripts/demo/validate.sh checks the built layer and a unit test runs each statement on sample records.'
logic_claims:
  - CLAIM-15EAF88EE2A8A371
  - CLAIM-613E39881528ED59
  - CLAIM-2C290FA62CD6A91A
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
semantic_source_hash: bccae387d4ec210e945fee9454993a566d22596858ed4897db49116419a5d4ad
semantic_inventory:
  - claim_key: CLAIM-15EAF88EE2A8A371
    claim_text: 'The demo''s telemetry collector scrubs visitors'' personal data before anything leaves it: the processor transform/scrub, which fails closed (error_mode propagate), runs in every logs, traces and metrics pipeline with nothing but the memory limiter before it'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 256
  - claim_key: CLAIM-613E39881528ED59
    claim_text: it drops the attributes that carry a client address, a session or token ID, a credential, Authorization, Cookie or another header, the user agent, or a question or prompt, and rewrites any IPv4 or IPv6 address, any openfief token prefix (ofs_, ofsess_, ofenroll_), Bearer or Basic credential, JSON web token and long token-like string, in the resource, scope, span, span event, log record and metric data point attributes, the log body and the span's name and status message
    role: descriptive
    status: missing
    span:
      start: 258
      end: 732
  - claim_key: CLAIM-2C290FA62CD6A91A
    claim_text: scripts/demo/validate.sh checks the built layer and a unit test runs each statement on sample records
    role: descriptive
    status: missing
    span:
      start: 734
      end: 835
id: REQ-demo-telemetry-scrub
type: req
---
The demo's telemetry collector scrubs visitors' personal data before anything leaves it: the processor transform/scrub, which fails closed (error_mode propagate), runs in every logs, traces and metrics pipeline with nothing but the memory limiter before it; it drops the attributes that carry a client address, a session or token ID, a credential, Authorization, Cookie or another header, the user agent, or a question or prompt, and rewrites any IPv4 or IPv6 address, any openfief token prefix (ofs_, ofsess_, ofenroll_), Bearer or Basic credential, JSON web token and long token-like string, in the resource, scope, span, span event, log record and metric data point attributes, the log body and the span's name and status message; scripts/demo/validate.sh checks the built layer and a unit test runs each statement on sample records.

## Context

The demo's logs, traces and metrics go to Grafana Cloud so the orchestrator can read them, and the demo is public: no visitor's address, session, token, header or question may leave it (GR-12, D-203). The core's own allow-list (D-172) comes first; the collector's scrub is what is left if it or the rule that nothing logs an address or a token fails. The author's condition for the shipping was that it is enforced in the collector, with a check.

## Source

> No visitor personal data leaves (hard condition), enforced in the collector: processors drop or scrub any attribute or log field that is a client IP, session IDs, token IDs and Authorization or Cookie headers; add a guard for attributes named like question, q or prompt; add a check that the scrub processors are in every pipeline before the exporter (the design session's demo operations brief, 2026-11-19).