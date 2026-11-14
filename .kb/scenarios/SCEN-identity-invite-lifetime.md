---
title: A viewer's invitation lasts an hour by default
status: active
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
id: SCEN-identity-invite-lifetime
type: scenario
---
Given an owner session with a recent passkey check, when the owner invites a viewer without choosing a lifetime the link works for an hour; 15 minutes and 24 hours are accepted and more than 24 hours is refused; the link works once; the owner's own enrollment link still lasts 15 minutes.