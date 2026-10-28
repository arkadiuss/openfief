---
title: The egress proxy refuses the metadata address in any disguise
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:31:57.172Z'
id: SCEN-plugins-egress-metadata-refused
type: scenario
---
Given a granted hostname that resolves to 169.254.169.254, ::ffff:169.254.169.254, 64:ff9b::a9fe:a9fe or 10.0.0.1, when the proxy checks the address before dialing, then each is refused as a forbidden destination, while a public address such as 9.9.9.9 is allowed. Assumes the proxy checks every resolved address, not only the first.