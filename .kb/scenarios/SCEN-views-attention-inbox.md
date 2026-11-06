---
title: The Inbox's views list attention items with their group and scope
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:24.278Z'
id: SCEN-views-attention-inbox
type: scenario
---
Given open, acknowledged and resolved attention items of several kinds, when a read caller lists one view, optionally for one account, ISIN or connector instance, then only those items come, each with its group and the account, ISIN and instance it is about, with counts of open, acknowledged and done items; an item tied to an account the caller may not view is left out.