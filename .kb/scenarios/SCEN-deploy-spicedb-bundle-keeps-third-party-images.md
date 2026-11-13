---
title: The release bundle ships SpiceDB at the pinned version and digest, and the database file brings its Database and role
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:16.840Z'
id: SCEN-deploy-spicedb-bundle-keeps-third-party-images
type: scenario
---
Given the release bundle rendered from deploy/, then its install file runs SpiceDB and the credentials Job's image exactly as deploy/spicedb pins them, by tag and digest, while every image of ours is pinned to its signed digest; SpiceDB's tag is the Taskfile's SPICEDB_VERSION; a third party's image by tag alone is refused; and the database file holds the Cluster openfief-db with the managed role spicedb and the Database spicedb it owns, and nothing else.