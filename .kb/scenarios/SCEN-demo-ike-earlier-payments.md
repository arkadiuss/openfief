---
title: The demo enters its seeded IKE's earlier payments
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:36.645Z'
id: SCEN-demo-ike-earlier-payments
type: scenario
---
Given the demo dataset's bond account, an IKE seeded at go-live from opening lots and opening cash, when the dataset is generated, then its earlier payments are its bonds' cost plus its opening cash in PLN, which the loader enters, so the demo shows its liquidation value taxed above them rather than missing.