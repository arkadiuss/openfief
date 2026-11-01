---
title: The release bundle brings CloudNativePG and a minimal openfief-db cluster
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T21:51:51.438Z'
id: SCEN-tooling-release-bundle-database
type: scenario
---
Given the deploy directory and a CloudNativePG release manifest with its pinned sha256, when the release bundle is rendered, then it holds openfief-db-<version>.yaml with exactly the CloudNativePG Cluster openfief-db in namespace openfief with database and owner openfief and a header saying what the owner may edit, and cnpg-<version>.yaml byte for byte; a missing, misnamed or differently summed manifest, or a database other than the one the install reads, is refused and nothing is written; and the install guide names the pinned version and sha256.