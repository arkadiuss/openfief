---
title: After the first passkey the owner is walked through setup
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:45.327Z'
id: SCEN-ui-first-run-wizard-steps
type: scenario
---
Given an install with no account, when the owner opens the overview, then the UI sends them to the wizard's first step, the default sleeves with their forms; the next step lists the reported broker accounts to create accounts for and the connector instances to bind; the last points to the overview, Accounts and Sleeves and to the CLI for more plugins; forms posted in the wizard return to its step; a read session is not sent there.