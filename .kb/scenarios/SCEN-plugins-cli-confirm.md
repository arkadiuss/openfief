---
title: Nothing changes until the owner confirms
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T23:36:03.761Z'
id: SCEN-plugins-cli-confirm
type: scenario
---
Given an install or uninstall the CLI has shown, when the owner answers n, presses enter or gives no input, then nothing in the cluster changes and the command fails as not confirmed; when the owner answers y or yes, or passed --yes, the change is applied and reported.