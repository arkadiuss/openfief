---
title: The core streams an owner's statement to the account's one bound connector and imports it as an upload batch
status: active
sourceFile: internal/ingest/app/uploads_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T04:57:59.310Z'
id: TEST-ingest-app-uploads
type: test
verification_scope: unit
---
Use-case tests of Uploads.Upload over in-memory fakes: a file reaches the one instance bound to the account as batch b-1 whole, with its name reduced to a hint and a sniffed media type replacing the declared one, and is imported once with source upload, keys scoped to the instance, seeding only in opening mode, then recorded once. Refusals, each importing and recording nothing: no or read-scoped session, no mode, an empty file, a file one byte over the limit or over it in its first bytes, a file that breaks off, zero or two bound instances, unreadable bindings, the connector refusing, taking no files or unreachable, a delivery into another account; a failing history fails the upload after the import.