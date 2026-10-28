---
title: The DCO check must fail when a commit in a pull request has no Signed-off-by trailer naming its author.
status: open
text_ref: design-decisions:D-5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-5
  - lane:strict
semantic_text: The DCO check must fail when a commit in a pull request has no Signed-off-by trailer naming its author.
logic_claims:
  - CLAIM-D9E99063699F8A78
semantic_clauses:
  - The DCO check must fail when a commit in a pull request has no Signed-off-by trailer naming its author.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 85722768f2b0a67d10209f53e6a12f62d20605595c78fcead93c1eefd88cf358
semantic_inventory:
  - claim_key: CLAIM-D9E99063699F8A78
    claim_text: The DCO check must fail when a commit in a pull request has no Signed-off-by trailer naming its author
    role: normative
    status: modeled
    span:
      start: 0
      end: 102
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:04.168Z'
id: REQ-tooling-dco-signoff
type: req
---
The DCO check must fail when a commit in a pull request has no Signed-off-by trailer naming its author.

## Context

D-5 says contributions carry a DCO sign-off; CI enforces it on every pull request so an unsigned commit never reaches main. The check compares each sign-off with the commit author, as the DCO requires the author to certify the contribution. Requested in the SDLC/CI implementation brief; the CI wiring is a Claude Code proposal marked P in docs/dev/sdlc.md.

## Source

> The core is licensed AGPL-3.0, the plugin contracts and SDK Apache-2.0, and contributions carry a DCO sign-off.

docs/DESIGN.md (D-5)
