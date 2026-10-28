---
title: The read API is one protobuf service served as Connect and as REST under /v1
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:43:07.728Z'
id: SCEN-views-api-contract
type: scenario
---
Given the ReadService defined in proto/openfief/views/v1 with HTTP annotations, when a generated Connect client calls a method or a REST client sends the annotated GET under /v1, then both reach the same handler and get the same data: money as an object with a decimal string amount and a currency code, quantities as decimal strings, dates as YYYY-MM-DD, valuations with basis, source, original-currency value and reporting-currency value; and every REST response matches the OpenAPI document generated from the same definitions. An unknown account is not found (404), an invalid request invalid argument (400), any other failure internal (500) naming only the request ID.