---
title: The release bundle's database sets its own requests and memory limit
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:14:28.455Z'
id: SCEN-tooling-release-bundle-database-resources
type: scenario
---
Given the deploy directory, when the release bundle is rendered, then openfief-db-<version>.yaml's Cluster openfief-db requests 100m CPU and 256Mi of memory and is limited to 1Gi of memory, so the namespace's LimitRange defaults never size the database, and its header and the install guide tell the owner to size them for their machine and history.