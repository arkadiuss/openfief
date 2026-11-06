---
title: SQL naming another context's schema fails the architecture test
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:47.524Z'
id: SCEN-tooling-archtest-foreign-schema
type: scenario
---
Given a portfolio migration selecting from valuation.tax_wrappers and an ingest query string joining ledger.accounts, when the schemas rule runs, then both files are reported with the foreign schema; a comment naming valuation and valuation's query of river.river_job are not.
