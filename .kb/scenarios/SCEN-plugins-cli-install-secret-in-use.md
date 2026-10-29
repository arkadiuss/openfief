---
title: An install whose Secret another instance names is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:09.929Z'
id: SCEN-plugins-cli-install-secret-in-use
type: scenario
---
Given an installed instance whose grant names the Secret sample-b-credentials, when the owner confirms an install whose grant would name that same Secret, then nothing is created and the install fails with ErrSecretInUse naming the Secret and the instance that holds it.