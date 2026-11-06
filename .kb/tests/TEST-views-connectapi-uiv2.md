---
title: The Inbox, activity, sync history and instrument calls over REST and Connect match the OpenAPI document
status: active
sourceFile: internal/views/connectapi/uiv2_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:23.070Z'
id: TEST-views-connectapi-uiv2
type: test
---
Serves the read API over in-memory ports and calls each new method over REST, validating every 200 answer against the generated OpenAPI document, and the activity also through a generated Connect client: attention groups, scopes and views; the activity's pages with their tokens and each kind's details; sync statuses and attempts with their pages and a 404; the instrument's detail, prices with the primary marked, holdings with lots and transactions with their account; bad parameters answered 400; enum numbers matching the domain's.