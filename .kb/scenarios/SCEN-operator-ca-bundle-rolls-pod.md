---
title: Creating, changing or deleting the CA bundle rolls the pod
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:00.841Z'
id: SCEN-operator-ca-bundle-rolls-pod
type: scenario
---
Given a running Plugin that expects an extra CA bundle and may read it, when the owner creates the Secret then the pod template gets the annotation openfief.io/ca-bundle-sha256 with the checksum of the Secret data in key order; when the owner changes it the checksum changes; when the owner deletes it the annotation goes; each change of the pod template replaces the pod. The operator reads the Secret again every minute, since it may only get that Secret by name.