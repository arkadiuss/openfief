---
title: The read API has no write endpoint and a script token reaches no owner service
status: active
tags:
  - acceptance:p10-7
origin:
  kind: agent
  recorded_at: '2026-11-08T04:07:00.987Z'
id: SCEN-e2e-read-api-no-write
type: scenario
---
Given the core and a script token, when the suite reads the OpenAPI document and the protobuf services and calls the core, then the document has only GET operations without a request body, every ReadService method is unary and declares no side effects, POST, PUT, PATCH and DELETE on every operation's path answer 405 with Allow: GET with the token and without it, and every method of every other service under proto/, the owner services among them, called over Connect with the script token, is refused: 404 on the browser port, permission_denied on the plugin port, and 404 for the connector contract, which the core doesn't serve.