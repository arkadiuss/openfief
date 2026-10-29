---
title: A connector names entries by broker account and lists the broker accounts it saw
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:01.485Z'
id: SCEN-connector-broker-accounts
type: scenario
---
Given a connector that knows no OpenFief account, when it delivers a statement, then each transaction and balance assertion carries broker_account and no account_id, and the delivery lists every broker account in broker_accounts; the core groups those entries apart and refuses one naming both accounts or a broker account with a space.