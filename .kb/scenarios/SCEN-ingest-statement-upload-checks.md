---
title: An upload is checked before a connector reads it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T04:10:57.321Z'
id: SCEN-ingest-statement-upload-checks
type: scenario
---
Given an upload from a read-scoped session, with no session, or of an empty file, when the owner's upload is handled, then it is refused before the connector is called and nothing is imported; and given a file named ../statements/flex.xml, the connector gets flex.xml and a media type sniffed from the first bytes.