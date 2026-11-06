---
title: The removed check pages lead to the Inbox
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:01.549Z'
id: SCEN-ui-no-check-pages
type: scenario
---
Given a browser that opens /attention, /tickers, /assertions, /price-checks or /transfers, when the UI answers, then it redirects for good to /inbox, which lists what those pages listed; each check's results stay on the account page.