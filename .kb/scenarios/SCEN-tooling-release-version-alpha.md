---
title: An alpha version is accepted only as YYYY.M.N-alpha.K
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:13.000Z'
id: SCEN-tooling-release-version-alpha
type: scenario
---
Given the alpha channel, when the release workflow checks a version, then 2026.11.0-alpha.1 and 2026.12.3-alpha.10 pass, while a stable or beta version, alpha.0, a K with a leading zero, a missing K, an upper case suffix, a v prefix and an alpha whose tag already exists are refused before anything is built; and an alpha version is refused on the beta and stable channels.