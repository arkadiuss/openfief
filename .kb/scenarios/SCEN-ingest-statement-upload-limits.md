---
title: An uploaded statement is checked before a connector reads it and stopped at 64 MiB
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:07.202Z'
id: SCEN-ingest-statement-upload-limits
type: scenario
---
Given an upload, when the caller is not the owner, no mode is given or the file is empty, then nothing reaches a connector; when the file passes 64 MiB, or breaks off, the call to the connector is stopped and nothing is imported; otherwise the connector gets only the cleaned last element of the file's name and a media type sniffed from the first 512 bytes in place of the one the browser declared.