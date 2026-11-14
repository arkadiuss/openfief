---
title: Prices sent again unchanged change nothing
status: active
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:58.470Z'
id: SCEN-valuation-inputs-report-changes
type: scenario
---
Given stored prices, FX rates or bond terms, when the same values are saved again, then the store reports none changed; when one differs, it reports it.

## Context

D-211: every change of state is an entry of the audit log, failing closed. Asked for in the S9 brief.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14