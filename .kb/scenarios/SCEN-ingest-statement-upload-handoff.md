---
title: The core streams an owner's statement to the account's bound connector and imports it as an upload batch
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:05.143Z'
id: SCEN-ingest-statement-upload-handoff
type: scenario
---
Given an owner-scoped session and an account bound to exactly one connector instance, when the owner uploads a statement file in opening or transactions mode, then the core streams it to that instance's ImportFile as a new batch, imports the delivery into that account only with source upload and keys scoped to the instance, seeding only in opening mode; uploading it again changes nothing, and zero or several bound instances, a refusal or a delivery into another account import nothing.