---
title: A plugin's install, upgrade and removal are entered once each as the operator's
status: active
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:46.073Z'
id: SCEN-identity-audit-plugin-installs
type: scenario
---
Given the core watching the Plugin resources, when a plugin appears, its generation changes or it is gone, then one entry of the operator names the instance and the generation, seeing the same again enters nothing, and an entry that can't be written is tried again.

## Context

D-211: every change of state is an entry of the audit log, failing closed.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S8 brief, from D-211, 2026-11-14