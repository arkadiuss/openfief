---
title: The Ask box evaluation page is generated from the bench's committed results
origin:
  kind: agent
  ref: 'AB9 brief: the Ask box evaluation article and the Ask box showing who answered, with the author''s D-222 (2026-11-15)'
  recorded_at: '2026-11-15T09:15:01.226Z'
status: active
id: SCEN-askeval-results-page
type: scenario
---
Given the trimmed result files committed under internal/views/ui/evalresults, one per bench run, when docgen builds the reference pages, then reference/ask-eval lists each model and host's latest full run with its verdict, failing cases, reliability, cost, date and configuration hash, then every run newest first with the bench's report schema, a run of some cases is marked diagnostic and carries no verdict, a model no call answered says so, and a run imported from a report.json keeps only those fields and never an answer or the prompt.