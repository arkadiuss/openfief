---
title: A connector calls its configured https base URL and nothing else
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:01.169Z'
id: SCEN-sdk-endpoint-base-url-https
type: scenario
---
Given a connector built on the Go SDK with a default base URL, when the owner configures no base URL then it calls the default; when the owner configures an https base URL with a port or path then it calls that one; when the configured base URL is plain http or carries a user, query or fragment then the endpoint is refused; and a request or redirect to another host or path outside the base URL is refused before anything is sent.