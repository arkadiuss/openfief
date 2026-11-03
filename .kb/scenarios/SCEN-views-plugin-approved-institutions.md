---
title: The plugins page shows approved and unapproved institutions
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:29.746Z'
id: SCEN-views-plugin-approved-institutions
type: scenario
---
Given a connector whose manifest declares two institutions and whose grant approves one, when the read API lists the plugins, then it carries both the declared institutions and approved_institutions, and the UI's plugins page says it books for the approved one only and that the other is not approved.