---
title: The treasury-bonds plugin maps lots and syncs index readings for the core only, reaching only GUS and NBP
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:02.406Z'
id: SCEN-plugins-treasury-bonds-connector
type: scenario
---
Given the plugin serving the connector contract behind the SDK's core-token check, when the core calls ImportBondLots, then it maps the lots into its configured account as of today in Warsaw, or answers INVALID_ARGUMENT naming the rule; when the core calls Sync, then it delivers the index readings after the watermark; a call without the core's token is refused; its manifest asks for the connector capability and egress to api-sdp.stat.gov.pl and static.nbp.pl only, with no secret.