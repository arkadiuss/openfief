---
title: A broker balance that differs is stored as flagged and the import is kept
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:27.203Z'
id: SCEN-ledger-app-assertion-flag
type: scenario
---
Given an import of a 1000 USD deposit with a broker balance assertion of 999 USD for that date, when it is imported, then the deposit is appended and the assertion is stored as flagged with the derived 1000; when a corrected 999 deposit is imported later, then every stored assertion of the account is checked again and the 999 assertion becomes matched, while a later one that now differs becomes flagged.