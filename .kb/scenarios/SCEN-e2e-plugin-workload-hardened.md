---
title: An operator-deployed plugin runs hardened with projected tokens only
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:42.680Z'
id: SCEN-e2e-plugin-workload-hardened
type: scenario
---
Given the signed plugin running, then its pod uses ServiceAccount plugin-<name> without an API token or service links, runs as non-root with RuntimeDefault seccomp, no privilege escalation, a read-only root filesystem and all capabilities dropped, and holds only the openfief-core and openfief-egress projected tokens.