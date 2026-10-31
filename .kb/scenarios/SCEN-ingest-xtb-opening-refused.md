---
title: An opening batch of an xStation export is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:01.774Z'
id: SCEN-ingest-xtb-opening-refused
type: scenario
---
Given an xStation export, when the core hands it to the xtb connector in opening mode, then the connector refuses it with the reason, so the owner uploads the whole history in transactions mode.