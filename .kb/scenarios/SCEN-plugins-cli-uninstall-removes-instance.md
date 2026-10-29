---
title: Uninstall removes the instance's grant, Plugin and its own Secrets
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:43:25.003Z'
id: SCEN-plugins-cli-uninstall-removes-instance
type: scenario
---
Given two instances of one image, sample-a and sample-b, which both name sample-credentials, and an unrelated instance, when the owner uninstalls sample-a and then sample-b, then each loses its grant and Plugin, sample-credentials is kept on the first uninstall and deleted on the second, the unrelated instance is untouched, and uninstalling sample-a again fails as not installed. Ledger data is never part of an uninstall.