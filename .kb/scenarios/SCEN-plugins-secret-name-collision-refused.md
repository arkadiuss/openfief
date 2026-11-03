---
title: An instance whose Secret name another instance holds is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T08:22:10.497Z'
id: SCEN-plugins-secret-name-collision-refused
type: scenario
---
Given instance sample-b mounting sample-b-credentials, when the owner upgrades instance sample to a version whose role is b-credentials, then PrepareUpgrade and Upgrade refuse it with ErrSecretInUse and the installed version stays; and given a running clash-a, when a Plugin clash whose role a-token names clash-a-token is applied with its own grant, then the operator runs nothing for it and says SecretHeld until clash-a is gone.