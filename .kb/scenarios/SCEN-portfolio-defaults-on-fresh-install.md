---
title: An install starts with the default sleeves; a started one gets only missing kinds
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:47.857Z'
id: SCEN-portfolio-defaults-on-fresh-install
type: scenario
---
Given an empty ledger, when the defaults are ensured, then the six built-in kinds and the sleeves bonds, core and satellite are stored; when they are ensured again after the owner renamed a kind and moved a sleeve, both changes are kept; given a ledger that already holds a sleeve of its own, such as the demo dataset's, ensuring the defaults stores the missing kinds and no default sleeve.