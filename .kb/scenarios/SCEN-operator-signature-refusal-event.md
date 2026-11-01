---
title: A refused signature is recorded once as a Warning event
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T13:51:04.784Z'
id: SCEN-operator-signature-refusal-event
type: scenario
---
Given a Plugin with a grant whose image the verifier refuses, when the operator reconciles it twice, then exactly one Warning event names the refusal reason and message; when the image changes to one the verifier accepts, no event is recorded; when it changes back, one more Warning event is recorded. Runs against envtest's API server with a fake event recorder.