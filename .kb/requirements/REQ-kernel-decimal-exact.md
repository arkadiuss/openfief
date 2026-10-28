---
title: Arithmetic on amounts, quantities and rates must be exact, and a result must be rounded only when the caller asks for it with an explicit number of places and rounding mode.
status: open
text_ref: AGENTS.md#hard-rules
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:agents-md-hard-rules
  - lane:strict
semantic_text: Arithmetic on amounts, quantities and rates must be exact, and a result must be rounded only when the caller asks for it with an explicit number of places and rounding mode.
logic_claims:
  - CLAIM-A325A33C946A4E8D
semantic_clauses:
  - Arithmetic on amounts, quantities and rates must be exact, and a result must be rounded only when the caller asks for it with an explicit number of places and rounding mode.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9fc668f50fca6c3b347a610c0bc53cb9976e11a3edd0b8025d153e199114d4e0
semantic_inventory:
  - claim_key: CLAIM-A325A33C946A4E8D
    claim_text: Arithmetic on amounts, quantities and rates must be exact, and a result must be rounded only when the caller asks for it with an explicit number of places and rounding mode
    role: normative
    status: modeled
    span:
      start: 0
      end: 172
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:45.544Z'
id: REQ-kernel-decimal-exact
type: req
---
Arithmetic on amounts, quantities and rates must be exact, and a result must be rounded only when the caller asks for it with an explicit number of places and rounding mode.

## Context

The author bans float64 for money, and the code organization proposal extends that to prices, rates and quantities with decimal-backed kernel types. Rounding is the caller's choice because tax rules, NBP rates and broker statements may round differently; the kernel never picks a mode or a number of places on its own. The kernel uses shopspring/decimal, which is arbitrary-precision, so add, subtract and multiply never lose digits. Modeled by Claude Code for the M0 foundation task (shared kernel); the author has not reviewed it yet.

## Source

> Domain logic is pure functions with table tests. No `float64` for money.

AGENTS.md#hard-rules

> Never `float64` for amounts, prices, rates or quantities. Use the kernel's decimal-backed types, and every amount carries its currency.

docs/dev/code-organization.md#rules
