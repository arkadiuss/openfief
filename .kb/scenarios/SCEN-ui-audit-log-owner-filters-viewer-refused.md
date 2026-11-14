---
title: The owner filters the audit log by actor, action, outcome and date, and a viewer is refused
status: active
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:23:00.435Z'
id: SCEN-ui-audit-log-owner-filters-viewer-refused
type: scenario
---
Given Settings' Audit log, when the owner filters by actor, action, outcome and a range of days, then the core is asked for exactly that and the page shows it; when a viewer, a read session or the demo's session asks, then the page is refused.

## Context

D-211: every change of state is an entry of the audit log, failing closed. Asked for in the S9 brief.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14