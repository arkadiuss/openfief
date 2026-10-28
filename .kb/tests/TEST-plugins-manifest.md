---
title: Manifest validation reports every invalid or duplicate field
status: active
sourceFile: internal/plugins/manifest_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:05.384Z'
id: TEST-plugins-manifest
type: test
---
Table tests start from a valid synthetic connector manifest, change one aspect per case and compare the full list of field errors (field name plus sentinel) returned by Manifest.Validate: no capability, an unknown or duplicate capability, wildcard, IP-literal, zero-port and duplicate egress hosts, bad Secret names, relative, unclean, reserved and nested mount paths, bad or duplicate data tiers and a config schema that is not JSON. Requested is checked to return sorted permissions with Secrets by name. A false pass would be a validator that stops at the first error; the expected lists hold several errors each.