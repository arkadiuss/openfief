---
title: The operator's trust comes only from CI, the catalog and the owner's list
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:29:27.872Z'
id: SCEN-operator-trust-config
type: scenario
---
Given the operator starting, when no owner signers file is mounted then it trusts only OpenFief's CI and the embedded catalog; when the file lists signers then they are added; when the file is invalid or has an unknown field then the operator does not start.