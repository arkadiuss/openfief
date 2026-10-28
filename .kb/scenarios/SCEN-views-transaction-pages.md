---
title: A client reads every current transaction exactly once, page by page
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:09.822Z'
id: SCEN-views-transaction-pages
type: scenario
---
Given an account's current transactions, when a client asks for a page of a given size and then passes each page's cursor to get the next, then the pages follow date then source then external ID order, never repeat or skip a transaction, and the last page has no cursor. A size outside 1 to 500 or a cursor the API did not write is refused.