---
title: The egress allow-list allows only exact host and port matches
status: active
sourceFile: internal/plugins/allowlist_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:00.615Z'
id: TEST-plugins-allow-list
type: test
---
A table test asks a synthetic allow-list about destinations: listed hosts on their listed ports are allowed, also in upper case or with one trailing dot; another port, an unlisted host, a subdomain or parent of a listed host, a suffix trick, an IP literal, two trailing dots, a Kelvin-sign spelling that Unicode lower-cases to k, and anything on an empty list are refused. A false pass would be suffix or Unicode-folded matching; the subdomain, suffix and Kelvin cases catch them.