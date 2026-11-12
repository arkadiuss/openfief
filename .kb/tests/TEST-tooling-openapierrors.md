---
title: tools/openapierrors adds each operation's error responses from the rules and fails on what it can't resolve
status: active
sourceFile: tools/openapierrors/errors_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-08T05:43:35.321Z'
id: TEST-tooling-openapierrors
type: test
---
Table tests over a miniature generated document with one operation with a path parameter and one without: every operation gets the rules' every-statuses, the one with a path parameter its pathParameters statuses, an operation named by operationId its own, which replaces a path parameter's, each as a reference to components.responses in status order with default last and the generator's formatting kept byte for byte; no rules leave the document as it was. It fails on a response the components lack, an operation the document lacks, a status the generator already declares, a document without paths or an operation without responses, and malformed YAML. parseRules refuses an unknown key and a status that is no 4xx or 5xx. TestRun rewrites a document file in place and refuses missing flags, extra arguments, missing files and malformed rules.