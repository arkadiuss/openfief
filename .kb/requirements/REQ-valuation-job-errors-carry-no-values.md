---
title: The revaluation workers must hand River a failed job's error named by the job and by the known errors in its chain, or else its innermost error's type, so that neither River's log nor river_job.errors keeps an amount, a holding, an ISIN or an account number.
status: open
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
semantic_text: The revaluation workers must hand River a failed job's error named by the job and by the known errors in its chain, or else its innermost error's type, so that neither River's log nor river_job.errors keeps an amount, a holding, an ISIN or an account number.
logic_claims:
  - CLAIM-62D71C1661715CD6
semantic_clauses:
  - The revaluation workers must hand River a failed job's error named by the job and by the known errors in its chain, or else its innermost error's type, so that neither River's log nor river_job.errors keeps an amount, a holding, an ISIN or an account number.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a4f149225b0b8ca07354b3181023ee6a5ea97a7c84447770c4b4530f26e52865
semantic_inventory:
  - claim_key: CLAIM-62D71C1661715CD6
    claim_text: The revaluation workers must hand River a failed job's error named by the job and by the known errors in its chain, or else its innermost error's type, so that neither River's log nor river_job.errors keeps an amount, a holding, an ISIN or an account number
    role: normative
    status: modeled
    span:
      start: 0
      end: 257
origin:
  kind: agent
  recorded_at: '2026-11-03T15:15:57.996Z'
id: REQ-valuation-job-errors-carry-no-values
type: req
---
The revaluation workers must hand River a failed job's error named by the job and by the known errors in its chain, or else its innermost error's type, so that neither River's log nor river_job.errors keeps an amount, a holding, an ISIN or an account number.

## Context

GR-12 keeps amounts, holdings, account numbers and tokens out of logs. River logs a failed job's error in full and stores it in river_job.errors, and a revaluation's errors can quote a lot's cost and quantity (valuation.ErrBondLot), an oversell's quantities or a transaction's sums. The review before the alpha tag (P1d) found it; the workers return a kernel.LogSafe-style JobError instead of wrapping River's logger, so the stored errors are covered too.

## Source

> Never log amounts, holdings, account numbers or tokens. Log internal IDs only (GR-12, UC-0).

docs/dev/code-organization.md#rules
