---
title: The Audit log page warns while entries wait to enter the log
status: active
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:06.337Z'
id: SCEN-ui-audit-log-behind-banner
type: scenario
---
Given the owner on Settings' Audit log, when the attention item audit-log-behind is open, then the page shows a banner with the item's count of changes waiting and that their entries will appear with the time of the change; with no such item, or a core that can't list items, no banner.

## Context

The banner is the same state as the Inbox's item, read through the attention items of the read API, so a core that can't list them shows the page without it. It follows the design system's alert component.

## Source

> AUDIT-LOG-BEHIND ATTENTION ITEM: after an audit-entry job has failed 5 attempts (keep River retrying indefinitely; never drop), raise attention item kind audit-log-behind in the data coming in group next to failing syncs (D-196). Cleared when the backlog is empty. Same state as a banner on Settings, Audit log.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18