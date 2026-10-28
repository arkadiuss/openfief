---
title: task dns:site must point <DEMO_DOMAIN> at GitHub Pages with DNS-only A and AAAA records to exactly GitHub's published Pages addresses and www.<DEMO_DOMAIN> with a DNS-only CNAME to <owner>.github.io, and with GITHUB_PAGES_CHALLENGE set a TXT record _github-pages-challenge-<owner>.<DEMO_DOMAIN>, each with the comment managed-by=openfief-site; it must create, update or delete only records with that comment, delete records with the retired comment managed-by=openfief-docs, never create a docs record, refuse before any change when another record of the same name and type or a conflicting CNAME exists, and task dns:site:plan must show the changes without making them; the Cloudflare token is handled as for the demo.
status: open
text_ref: docs/dev/demo.md#dns
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-dns
  - lane:strict
semantic_text: task dns:site must point <DEMO_DOMAIN> at GitHub Pages with DNS-only A and AAAA records to exactly GitHub's published Pages addresses and www.<DEMO_DOMAIN> with a DNS-only CNAME to <owner>.github.io, and with GITHUB_PAGES_CHALLENGE set a TXT record _github-pages-challenge-<owner>.<DEMO_DOMAIN>, each with the comment managed-by=openfief-site; it must create, update or delete only records with that comment, delete records with the retired comment managed-by=openfief-docs, never create a docs record, refuse before any change when another record of the same name and type or a conflicting CNAME exists, and task dns:site:plan must show the changes without making them; the Cloudflare token is handled as for the demo.
logic_claims:
  - CLAIM-A05EEF8217267293
semantic_clauses:
  - task dns:site must point <DEMO_DOMAIN> at GitHub Pages with DNS-only A and AAAA records to exactly GitHub's published Pages addresses and www.<DEMO_DOMAIN> with a DNS-only CNAME to <owner>.github.io, and with GITHUB_PAGES_CHALLENGE set a TXT record _github-pages-challenge-<owner>.<DEMO_DOMAIN>, each with the comment managed-by=openfief-site; it must create, update or delete only records with that comment, delete records with the retired comment managed-by=openfief-docs, never create a docs record, refuse before any change when another record of the same name and type or a conflicting CNAME exists, and task dns:site:plan must show the changes without making them; the Cloudflare token is handled as for the demo.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ffec167d111471140e1596fd5be79d5fee8da1cac6395b47ed6a55d94520e412
semantic_inventory:
  - claim_key: CLAIM-A05EEF8217267293
    claim_text: task dns:site must point <DEMO_DOMAIN> at GitHub Pages with DNS-only A and AAAA records to exactly GitHub's published Pages addresses and www.<DEMO_DOMAIN> with a DNS-only CNAME to <owner>.github.io, and with GITHUB_PAGES_CHALLENGE set a TXT record _github-pages-challenge-<owner>.<DEMO_DOMAIN>, each with the comment managed-by=openfief-site; it must create, update or delete only records with that comment, delete records with the retired comment managed-by=openfief-docs, never create a docs record, refuse before any change when another record of the same name and type or a conflicting CNAME exists, and task dns:site:plan must show the changes without making them; the Cloudflare token is handled as for the demo
    role: normative
    status: modeled
    span:
      start: 0
      end: 718
origin:
  kind: agent
  recorded_at: '2026-10-28T15:29:35.486Z'
id: REQ-tooling-site-dns
type: req
---
task dns:site must point <DEMO_DOMAIN> at GitHub Pages with DNS-only A and AAAA records to exactly GitHub's published Pages addresses and www.<DEMO_DOMAIN> with a DNS-only CNAME to <owner>.github.io, and with GITHUB_PAGES_CHALLENGE set a TXT record _github-pages-challenge-<owner>.<DEMO_DOMAIN>, each with the comment managed-by=openfief-site; it must create, update or delete only records with that comment, delete records with the retired comment managed-by=openfief-docs, never create a docs record, refuse before any change when another record of the same name and type or a conflicting CNAME exists, and task dns:site:plan must show the changes without making them; the Cloudflare token is handled as for the demo.

## Why this changed

The design agent moved the website to the apex: a landing page at https://openfief.dev/ and the docs at https://openfief.dev/docs/, served by GitHub Pages from Actions, so docs.openfief.dev is no longer used. The orchestrating agent asked to replace task dns:docs with task dns:site and task dns:site:plan for the apex and www, and to delete the old docs record wherever it remains, because a CNAME to github.io that no Pages site claims is a subdomain-takeover risk. The old requirement's single CNAME at docs.<DEMO_DOMAIN> would recreate exactly that record.

## Context

GitHub Pages issues its own certificate, so the records are not proxied. The apex cannot be a CNAME, so it gets GitHub's four IPv4 and four IPv6 addresses, reconciled as a set. The TXT record verifies the domain for the owner's GitHub account, so no other account can claim it; its value comes from the account's Pages settings and is optional. A docs.<DEMO_DOMAIN> redirect would need a Cloudflare Single Redirect rule, which the token's permissions don't allow, so it is not automated. Separate comments keep the demo and site records from ever touching each other. Modeled by Claude Code. The author has not reviewed this requirement yet.

## Source

> Replace `task dns:docs` with `task dns:site` and `task dns:site:plan` (read-only), which ensure, all DNS-only (proxied=false), with comment `managed-by=openfief-site` ... cleanup: delete any record still tagged `managed-by=openfief-docs` (the retired docs record), and never create docs.* any more.

Orchestrator brief for impl/site-dns; docs/dev/demo.md#dns
