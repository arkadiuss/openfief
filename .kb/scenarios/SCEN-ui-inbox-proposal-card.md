---
title: The Inbox shows a proposal with its figures to the owner alone
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:35.319Z'
id: SCEN-ui-inbox-proposal-card
type: scenario
---
Given a pending proposal, when the owner opens the Inbox then each change shows its target, weight now and proposed, weight held, drift and reason with Apply and Dismiss; a read session sees only that it waits for the owner; a decided proposal shows what was decided and no buttons.