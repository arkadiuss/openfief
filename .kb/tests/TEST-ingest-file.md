---
title: A statement file's name reaches a connector only as a cleaned last path element
status: active
sourceFile: internal/ingest/file_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T04:57:57.650Z'
id: TEST-ingest-file
type: test
verification_scope: unit
---
Table test of ingest.FileNameHint: Unix and Windows paths keep only their last element, a trailing slash, '.' and '..' become empty, control characters and invalid UTF-8 are dropped, spaces trimmed, a Unicode name kept, and a long name cut to 255 bytes on a character boundary. A false pass would need a path separator to survive, which the Unix and Windows cases rule out.