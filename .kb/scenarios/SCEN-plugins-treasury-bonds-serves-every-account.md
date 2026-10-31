---
title: A bond lot goes to the connector with its account
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:22.992Z'
id: SCEN-plugins-treasury-bonds-serves-every-account
type: scenario
---
Given a treasury-bonds instance bound to several bond accounts, when the owner enters a lot on one of them, then the core hands the lot to the instance with that account and the connector books it there; its config names no account, and a lot without one is refused.