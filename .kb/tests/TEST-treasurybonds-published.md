---
title: The connector fills in the terms the owner left out, and the owner's terms win
status: active
sourceFile: internal/ingest/treasurybonds/published_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T08:05:58.735Z'
id: TEST-treasurybonds-published
type: test
---
TestCompleteTerms runs CompleteTerms on a synthetic lot left wholly, partly or not at all to the connector against a counting fake of the published data: published rates with the attribution whose processing names the fee's origin, the reviewed table's fee by sale month, no read and no attribution for the owner's terms, and ErrNotPublished and ErrNoReviewedFee for series the data or table don't cover. TestCompleteTermsReadsOnce checks one read per batch and that a failure to read is returned as it came.