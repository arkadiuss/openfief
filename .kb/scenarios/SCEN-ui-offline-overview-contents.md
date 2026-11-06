---
title: Only the overview's values and shares are handed over to keep, and kept encrypted
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T10:10:59.296Z'
id: SCEN-ui-offline-overview-contents
type: scenario
---
Given an owner or read session, when the overview page renders, then it carries in data-offline-overview the total's gross, net and liquidation values, each account's name, institution, sleeve and three values, the allocation's shares, the date and when the session ends, in the page's number format, and no account ID, transaction or lot; no other page or fragment carries anything to keep, and in the browser the copy is stored only as AES-GCM ciphertext beside a non-extractable key, with nothing in localStorage or a page in sessionStorage.