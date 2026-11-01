---
title: The core's stand-in for the attention store logs each item raised and cleared
status: active
sourceFile: cmd/core/unclassified_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T19:57:18.771Z'
id: TEST-core-unclassified
type: test
---
loggedAttention over a text slog handler raises and clears an instrument-unclassified item for a synthetic ISIN and the log holds the kind and subject of each, with no error; it stands in until the attention store (D-130) is merged.