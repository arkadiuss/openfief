---
title: Nothing changes until the owner confirms the upgrade
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T17:04:01.597Z'
id: SCEN-plugins-upgrade-applies-on-confirmation
type: scenario
---
Given the upgrade summary, when the owner answers anything but y or yes, then the Plugin and grant stay as they were; when the owner confirms, the Plugin's spec is replaced first and then the grant's, so the previous version keeps running until the grant approves the new one.