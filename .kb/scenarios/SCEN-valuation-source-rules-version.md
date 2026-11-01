---
title: Upgrading to one source per series revalues everything
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:02.575Z'
id: SCEN-valuation-source-rules-version
type: scenario
---
Given values stored under the earlier rule, the highest-ranked source with a value on each date, when the core starts with the source rules version 2, then the recorded rules version differs and every account is revalued from its first transaction, each series taking its source on that first need.