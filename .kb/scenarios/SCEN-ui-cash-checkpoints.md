---
title: The owner enters a cash checkpoint and books its difference from the account page
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:50:20.151Z'
id: SCEN-ui-cash-checkpoints
type: scenario
---
Given an account whose cash checkpoint differs from the derived cash, when the owner opens the account page from the attention item, then the page lists the checkpoint with the cash shown, derived and the difference, and a Book as a deposit or Book as a withdrawal button that books it; the Enter cash checkpoint form takes a date, an amount and a currency; a read session sees the list without the button or the form.