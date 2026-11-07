---
title: The ministry's spreadsheet is discovered and read, gov.pl first
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:03.932Z'
id: SCEN-valuation-mfbonds-published-rates
type: scenario
---
Given gov.pl's page linking the spreadsheet and dane.gov.pl's resource list naming it, when the published data is read, then the link is found on the page and the rates read with the attribution, and whenever anything about gov.pl fails the dane.gov.pl copy is read instead; links to other hosts, oversized bodies and damaged files are refused.