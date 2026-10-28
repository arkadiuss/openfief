---
title: A sleeve kind can be renamed but its default role never changes
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:26.692Z'
id: SCEN-portfolio-kind-role-fixed
type: scenario
---
Given the built-in kind retirement, which is defensive, when the owner renames it to IKE and IKZE, then the rename is accepted; when the owner saves bonds as offensive or agent-pool as defensive, then it is refused with a kind role change error, because sleeves of that kind were checked against its old default role.