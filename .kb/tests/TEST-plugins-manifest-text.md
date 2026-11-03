---
title: A manifest with a control or bidirectional character anywhere is refused
status: active
sourceFile: internal/plugins/text_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T09:18:00.223Z'
id: TEST-plugins-manifest-text
type: test
---
TestManifestRefusesUnsafeText validates samplePlugin's manifest with one string changed per case: the review's payload in a schema property name and a cluster service description, a C1 control in a schema title, a right-to-left override in an enum value, a bidi isolate in a description, invalid UTF-8 in the schema, an escape in a mount path and a tab in a config field name must fail with ErrUnsafeText, and plain text with titles and an em dash must pass. TestSafeText tables SafeText over C0, DEL, C1, overrides, isolates and UTF-8.