---
title: Egress hosts are canonical hostnames with a port, never IP literals or wildcards
status: active
sourceFile: internal/plugins/egress_host_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:01.368Z'
id: TEST-plugins-egress-host
type: test
---
Table tests in package plugins_test call ParseEgressHost with synthetic example.com and example.org names. They accept a bare hostname as port 443, an explicit port, an IDNA A-label and a 63-character label, and reject upper case, raw Unicode, single labels such as localhost, trailing dots, empty labels, edge hyphens, reserved double hyphens, underscores, over-long labels and names, IPv4 and IPv6 literals with and without ports, short IPv4 forms such as 127.1, wildcards and bad ports, each with its sentinel error compared through cmpopts.EquateErrors. A false pass would be an IP literal or wildcard accepted as a hostname; each has its own case.