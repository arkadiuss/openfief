---
title: A connector's signed manifest must declare the institutions it books for, and the core must refuse, whole, a delivery that names an institution the running version's manifest doesn't declare, importing and storing nothing of it.
status: open
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
semantic_text: A connector's signed manifest must declare the institutions it books for, and the core must refuse, whole, a delivery that names an institution the running version's manifest doesn't declare, importing and storing nothing of it.
logic_claims:
  - CLAIM-C0B7D02B561F51CD
semantic_clauses:
  - A connector's signed manifest must declare the institutions it books for, and the core must refuse, whole, a delivery that names an institution the running version's manifest doesn't declare, importing and storing nothing of it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1663ea81f419df7a2277c39664321e47aa14784845f7364750fa26196708e4a2
semantic_inventory:
  - claim_key: CLAIM-C0B7D02B561F51CD
    claim_text: A connector's signed manifest must declare the institutions it books for, and the core must refuse, whole, a delivery that names an institution the running version's manifest doesn't declare, importing and storing nothing of it
    role: normative
    status: modeled
    span:
      start: 0
      end: 227
origin:
  kind: agent
  recorded_at: '2026-11-02T22:20:04.852Z'
id: REQ-ingest-delivery-declared-institution
type: req
---
A connector's signed manifest must declare the institutions it books for, and the core must refuse, whole, a delivery that names an institution the running version's manifest doesn't declare, importing and storing nothing of it.

## Context

The core trusted the institution a delivery named: it stamped the batch and the market-data sources with it, and the default ranking of sources puts eodhd, nbp and treasury-bonds first, breaking ties by instance name, so an instance named aaa claiming eodhd took EODHD's place. The declaration is signed with the image (D-42) and the install shows it. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p1-isolation-net), from the security review the author put before the alpha tag. No firm requirement for this part of D-182 existed in kibi on this branch when it was built.

## Source

> A plugin's Secrets are named so they can never collide with the platform's, and a connector books only for the institutions its signed manifest declares.

docs/DESIGN.md#D-182
