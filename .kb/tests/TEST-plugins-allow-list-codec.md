---
title: Egress allow-lists encode to and decode from ConfigMap data strictly
status: active
sourceFile: internal/plugins/allowlist_codec_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:47.443Z'
id: TEST-plugins-allow-list-codec
type: test
---
Table tests in package plugins_test call EncodeAllowLists and DecodeAllowLists with synthetic example.com plugins and hosts. Encoding sorts hosts, writes one host:port per line and gives a plugin with no hosts an empty value; it rejects invalid plugin names, duplicate plugins, wildcard and duplicate hosts. Decoding returns lists sorted by plugin and rejects keys that are not plugin names, hosts without a port, upper case, IP literals, blank lines, a missing final newline and duplicate hosts, each with its field and sentinel error. A round trip returns the input. A false pass would be a lenient decoder that guesses a port; the port-less case guards it.