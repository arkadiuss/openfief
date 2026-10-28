---
title: An account is stored only under a stored sleeve, and imports need a stored account
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:32.188Z'
id: SCEN-ledger-app-registry
type: scenario
---
Given a stored sleeve core, when an account in sleeve core is saved, then it is stored and listed; when an account names a sleeve that is not stored, then saving fails with an unknown sleeve error and nothing changes; when an import or a balance assertion names an account that is not stored, then the import fails with an unknown account error before anything is written.