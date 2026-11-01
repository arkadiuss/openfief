---
title: A config field's https URL gives one egress host that the grant approves and the proxy enforces
status: active
sourceFile: internal/plugins/config_host_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T15:09:56.135Z'
id: TEST-plugins-config-host
type: test
verification_scope: unit
---
Table tests over a synthetic view plugin whose manifest names config field url: ConfigHosts reads the host of an https URL with its port, lower-cased, and refuses plain http, IPv4 and IPv6 addresses, wildcards, single labels, a user part, port zero, non-strings and a config that is not an object; an absent, null or empty field gives none. The host is requested, compiled into the allow-list behind the proxy rule, and a moved host waits for approval.