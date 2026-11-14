---
title: The owner filters the audit log by outcome and by a set of actions
status: active
origin:
  kind: agent
  ref: 'S7 task brief: the Access page''s history and Settings'' audit log view (D-210, D-211), 2026-11-13'
  recorded_at: '2026-11-14T00:30:00.000Z'
id: SCEN-identity-audit-filter-outcome-actions
type: scenario
---
Given an owner session, ListAuditEntries selects entries by outcome, by any of a set of actions, or both with one action, at the domain, over Connect and in PostgreSQL; an unknown outcome or action in the set is an invalid argument; a viewer's session is the actor of their entries and comes with the ID of the person it is of.