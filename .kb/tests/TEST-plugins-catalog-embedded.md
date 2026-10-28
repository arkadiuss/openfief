---
title: The catalog released with the product parses and pins each digest once
status: active
sourceFile: internal/plugins/catalog/catalog_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:29:19.502Z'
id: TEST-plugins-catalog-embedded
type: test
---
Loads the catalog file embedded from internal/plugins/catalog/plugins.json, builds a trust policy from it and checks that no digest is pinned twice. A false pass would be a catalog that is never parsed at build time; the test parses the same embedded bytes the operator uses.