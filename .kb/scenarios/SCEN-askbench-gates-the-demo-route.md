---
title: The bench reads the demo route, reports who may join or must leave and fails when a member fails
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:44.274Z'
status: active
id: SCEN-askbench-gates-the-demo-route
type: scenario
---
Given the demo route's config in deploy/demo/model, when the bench runs with -demo-route and no model ids, then it asks the route's members with their hosts, reasoning, structured-output flag and prices, a member whose run fails its verdict or its reliability makes the bench exit 4 (after a failed safety property or an aborted budget) so the weekly workflow fails and the author is notified, the console and the summary name the labels eligible to join and the members with two consecutive failing runs, a run of some cases is not gated, and the route's files are never written.