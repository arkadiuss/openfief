---
title: A connector receives an imported file in order
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:48.922Z'
id: SCEN-sdk-go-receive-file
type: scenario
---
Given an ImportFile stream, when it starts with one header followed by chunks, then ReceiveFile returns the header and the chunks joined in order; when it has no header, a second header, or more bytes than the connector's limit, ReceiveFile refuses it.