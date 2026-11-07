---
title: ca-bundle unset deletes the bundle Secret and clears the stamp
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T17:32:52.585Z'
id: SCEN-plugins-cli-ca-bundle-unset
type: scenario
---
Given an installed instance with a CA bundle set, when the owner runs openfief plugin ca-bundle unset <instance> and confirms then the CLI deletes the Secret <instance>--ca-bundle, only when it is labelled as the instance, and then clears the Plugin spec.caBundleSha256; declining changes nothing, and an instance that is not installed is refused.