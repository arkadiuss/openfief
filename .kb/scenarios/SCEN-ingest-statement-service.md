---
title: A view plugin streams the owner's statement to the core, which answers with the batch or the reason
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:11.021Z'
id: SCEN-ingest-statement-service
type: scenario
---
Given a stream from a view plugin whose first message is the header with the owner's origin, session, account, file name and mode, when the file's chunks follow, then the core runs the upload and answers the batch's counts with at most 100 ingest problems, or UNAUTHENTICATED, PERMISSION_DENIED, INVALID_ARGUMENT with the reason, RESOURCE_EXHAUSTED, FAILED_PRECONDITION, UNAVAILABLE or INTERNAL without the cause.