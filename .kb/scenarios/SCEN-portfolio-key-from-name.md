---
title: The owner's names become keys and IDs
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:50.732Z'
id: SCEN-portfolio-key-from-name
type: scenario
---
Given the names 'Emergency fund', 'Konto glowne' with Polish letters and '  IBKR -- main (USD)! ', when a key is made from each, then they are emergency-fund, konto-glowne and ibkr-main-usd, valid kind keys of at most 64 characters; a name of only punctuation fails with ErrNoKey.