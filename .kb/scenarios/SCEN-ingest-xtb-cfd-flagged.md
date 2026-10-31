---
title: XTB CFD rows are flagged
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:54.598Z'
id: SCEN-ingest-xtb-cfd-flagged
type: scenario
---
Given an export with open and closed CFD positions and a cash operation of a CFD category, when it is mapped, then each is an unsupported ingest problem naming its category and nothing of it is booked.