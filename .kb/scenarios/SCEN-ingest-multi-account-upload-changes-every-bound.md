---
title: A statement upload needs the right to change every account its instance may book
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:32.729Z'
id: SCEN-ingest-multi-account-upload-changes-every-bound
type: scenario
---
Given a connector instance bound to three accounts and an owner whose authorizer denies changing one of them, when the owner uploads a statement for another of them, then the upload is refused as not permitted before the connector is called, because a statement of several accounts may book any account bound to the instance (D-140, D-163); with every account allowed it is imported.