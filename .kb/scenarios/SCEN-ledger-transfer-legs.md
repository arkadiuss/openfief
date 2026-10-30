---
title: Only a withdrawal and a deposit that can be one transfer are linked
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T06:05:36.004Z'
id: SCEN-ledger-transfer-legs
type: scenario
---
Given a withdrawal of 100 PLN from one account on 2 March, when the owner links it with a deposit of 100 PLN into another account dated 3 March, the link is stored; with a deposit dated 20 March, into the same account, of another amount or currency, opening cash, or a key with no stored transaction, the decision is refused and nothing is stored. The legs' current versions decide.