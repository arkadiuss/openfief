---
title: A new CA bundle stamp rolls the pod
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:00.841Z'
id: SCEN-operator-ca-bundle-rolls-pod
type: scenario
---
Given a running Plugin that expects an extra CA bundle, when the owner stamps a checksum on its spec then the pod template gets the annotation openfief.io/ca-bundle-sha256 with exactly that checksum; when the owner stamps another one the annotation changes; when the owner clears it the annotation goes; each change of the pod template replaces the pod. The operator reads no Secret for it and requeues nothing on a timer.