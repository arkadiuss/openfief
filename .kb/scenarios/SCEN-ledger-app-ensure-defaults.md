---
title: Ensuring the default kinds and sleeves at install is idempotent and keeps the owner's changes
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:52:58.113Z'
id: SCEN-ledger-app-ensure-defaults
type: scenario
---
Given an empty install, when the defaults are ensured, then the six built-in kinds and the sleeves Bonds, Core and Satellite are stored; when they are ensured again, or by several processes at once, then nothing more is stored and no call fails; given the owner renamed a kind or moved a default sleeve to another kind, when the defaults are ensured again, then those changes are kept.