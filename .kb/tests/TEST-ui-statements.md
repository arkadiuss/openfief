---
title: The owner uploads a statement on an account's page, the UI streams it to the core and shows the import
status: active
sourceFile: internal/views/ui/statements_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T05:39:07.562Z'
id: TEST-ui-statements
type: test
verification_scope: integration
---
Handler tests of the UI against an in-memory Connect fake of the StatementService over HTTP/1.1: only an owner session sees the multipart Upload statement form, mode before file; a posted 300000-byte file reaches the core whole with the plugin's token, owner session, origin, account, file name and mode, and the page shows the batch's counts and listed ingest problems. Refusals: a connector refusal 400 with its reason, an unreachable connector 503, a read session 403, the file before the mode, an unknown mode and no file 400 without reaching the core, a body over 64 MiB 413 with the upload stopped. The template passes the no-inline-code check.