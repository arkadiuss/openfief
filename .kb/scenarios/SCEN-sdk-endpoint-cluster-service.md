---
title: A connector reaches its approved in-cluster service directly, over HTTPS verified for the base URL's host
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T23:25:45.888Z'
id: SCEN-sdk-endpoint-cluster-service
type: scenario
---
Given an instance whose grant approves an in-cluster service and whose base URL names the service in-cluster DNS name, when the connector calls the endpoint then every request goes straight to the service address outside the egress proxy, with the base URL host in TLS and the Host header, and its certificate is verified against the system roots plus the extra CA bundle; a certificate for another name or from an unknown CA is refused, and so is any request outside the base URL.