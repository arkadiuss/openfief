---
title: The chat client retries a 429 once, after the response's Retry-After (seconds or an HTTP date) capped at ten seconds, or after one second when there is none, and only when that wait ends before the call's deadline; a second 429 is an unavailable provider that carries status 429, and no other status is retried by the client.
status: open
tags:
  - askbench
origin:
  kind: agent
  ref: 'Design session decision: 429 handling, retried once honouring Retry-After (D-207), 2026-11-15'
  recorded_at: '2026-11-15T09:14:03.255Z'
semantic_text: The chat client retries a 429 once, after the response's Retry-After (seconds or an HTTP date) capped at ten seconds, or after one second when there is none, and only when that wait ends before the call's deadline; a second 429 is an unavailable provider that carries status 429, and no other status is retried by the client.
logic_claims:
  - CLAIM-93CA70BEF1011E6F
  - CLAIM-0BB12921B8C36CE9
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
required_evidence: unverified
semantic_source_hash: a8e3f898ee494e67c7d41e959baadbfd425af7fd824111091e6651b23c113f9d
semantic_inventory:
  - claim_key: CLAIM-93CA70BEF1011E6F
    claim_text: The chat client retries a 429 once, after the response's Retry-After (seconds or an HTTP date) capped at ten seconds, or after one second when there is none, and only when that wait ends before the call's deadline
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 213
  - claim_key: CLAIM-0BB12921B8C36CE9
    claim_text: a second 429 is an unavailable provider that carries status 429, and no other status is retried by the client
    role: descriptive
    status: ontology_gap
    span:
      start: 215
      end: 324
id: REQ-models-chat-client-retries-429-once
type: req
---
The chat client retries a 429 once, after the response's Retry-After (seconds or an HTTP date) capped at ten seconds, or after one second when there is none, and only when that wait ends before the call's deadline; a second 429 is an unavailable provider that carries status 429, and no other status is retried by the client.

## Context

The Ask box bench saw 429s from OpenRouter hosts and counted them as no failure of the host at all, and a rate-limited call was never tried again, in the bench or in production. D-207 makes a host's reliability a reason to keep it off the box, so a 429 is retried once, within the question's two minutes, and one that stays a 429 counts against the host. OpenRouter itself passes over a rate-limited pinned host for the next listed one when allow_fallbacks is false, so the client does not walk the hosts.

## Source

Design session decision, 2026-11-15: "a 429 is retried once, honouring Retry-After (seconds or HTTP date) capped at 10 s (no Retry-After: a short fixed wait, e.g. 1 s), staying inside the 2-minute question timeout."
