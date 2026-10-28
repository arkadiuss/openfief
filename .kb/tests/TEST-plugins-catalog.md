---
title: The catalog and owner signer documents are parsed strictly
status: active
sourceFile: internal/plugins/catalog_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:50:23.599Z'
id: TEST-plugins-catalog
type: test
---
Table tests parse synthetic catalog and owner-signer JSON: keyless and key publishers and signers are read with their kinds; a catalog publisher pinned by a pattern, an image by tag, an entry without a name, a publisher without an identity, unknown fields such as trustEverything or allowUnsigned, trailing data and YAML text are all refused with ErrInvalidCatalog or ErrInvalidSigner. A false pass would be a lenient decoder; the unknown-field cases cover it.