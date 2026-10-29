---
title: Transactions a removed instance brought in stay listed and are marked as coming from a removed instance
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:32.324Z'
id: SCEN-views-removed-instance-mark
type: scenario
---
Given transactions whose key source is plugin:<instance>:<source>, when the read API lists them, then each names its connector instance and is marked removed when no Plugin of that name is installed any more; reinstalling the instance under the same name clears the mark; while the installed plugins are unknown nothing is marked; transactions not from a connector instance have no instance.