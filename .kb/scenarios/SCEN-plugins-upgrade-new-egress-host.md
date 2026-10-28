---
title: An upgrade that adds an egress host lists it as needing approval
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:15.682Z'
id: SCEN-plugins-upgrade-new-egress-host
type: scenario
---
Given a Grant that approves everything a plugin's current version requests, when a new version with a new image digest also requests new.example.net:443 and api.example.com:8443, then NeedsApproval returns exactly those two egress hosts and nothing else, so the operator keeps the old version until the owner approves them; and a new version that only changes its image digest needs no approval.