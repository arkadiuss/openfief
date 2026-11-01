---
title: Read access sees an unmapped broker account's item without its number
status: active
sourceFile: internal/attention/readsubject_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T10:05:30.064Z'
id: TEST-attention-read-subject
type: test
---
TestReadSubject runs attention.ReadSubject over a table of kinds and subjects: a broker-account-unmapped subject of instance and broker account comes back as the instance alone, and a ticker, batch and instrument subject come back unchanged, so read access never sees a broker account number (D-121, D-137).