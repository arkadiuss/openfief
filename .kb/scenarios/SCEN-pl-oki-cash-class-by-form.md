---
title: An OKI account's form decides its PLN cash's exemption class
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:01.539Z'
id: SCEN-pl-oki-cash-class-by-form
type: scenario
---
Given PLN cash on OKI accounts, when the Polish pack classifies it, then on a bank account or treasury bond account it is a savings asset, on a brokerage account or fund register an investment asset, on an account without a form other assets, and cash in another currency is other assets whatever the form; and given a treasury bond account's days, the views count its PLN cash as savings.