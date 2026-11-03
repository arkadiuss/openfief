---
title: The install warns about an institution another instance books for
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:23.545Z'
id: SCEN-plugins-install-warns-shared-institution
type: scenario
---
Given an installed instance that books for an institution, when the owner installs another connector that declares it, or upgrades one so that its institutions change, then the summary lists the institutions and warns, naming the other instance, before the owner confirms; with no such institution it warns about nothing.