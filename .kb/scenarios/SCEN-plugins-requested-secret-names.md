---
title: Each Secret role a manifest names becomes the instance's own Secret name
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:06.738Z'
id: SCEN-plugins-requested-secret-names
type: scenario
---
Given a manifest naming the Secret role flex-token, when the image is installed as instances ibkr-joint and ibkr-solo, then the Plugin of ibkr-joint requests and mounts the Secret ibkr-joint-flex-token and the Plugin of ibkr-solo requests ibkr-solo-flex-token; a role that is not a DNS label is refused by name.