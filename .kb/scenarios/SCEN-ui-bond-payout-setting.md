---
title: The account form offers the payout setting of a bond account
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:30.223Z'
id: SCEN-ui-bond-payout-setting
type: scenario
---
Given the owner on the accounts page, when the account is at the treasury bond registry, then its change form offers 'Stay as the account's cash' and 'Paid out' and hands the choice to the core; a broker account's form doesn't offer it, and a form without the field keeps the account's setting.