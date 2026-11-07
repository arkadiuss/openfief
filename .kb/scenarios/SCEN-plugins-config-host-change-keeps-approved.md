---
title: A base URL moved to an unapproved host keeps the old approved allow-list
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:01.820Z'
id: SCEN-plugins-config-host-change-keeps-approved
type: scenario
---
Given a connector running with a base URL on a host its grant approves, when its Plugin is changed to a base URL on another host while the grant still approves only the old one, then the running version keeps running with the old approved egress hosts, the new host is pending approval and never reaches the allow-list, and the upgrade diff lists the new host as added and the old one as removed.