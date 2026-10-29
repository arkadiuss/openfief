---
title: The allow-list holds only the host request and grant share
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:47.298Z'
id: SCEN-e2e-plugin-egress-allow-list
type: scenario
---
Given a manifest requesting api.nbp.pl and a grant approving api.nbp.pl and example.org, when the plugin runs, then its entry in openfief-egress-allowlist is exactly api.nbp.pl:443 and status.effective lists only that host.