---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Tracing

The owner asked to "add tracing and make pushing to otel configurable". No decision covers observability, so the narrowest version is built ([Send traces to a collector](guides/tracing.md)): OpenTelemetry tracing in the core, the operator and the Go SDK, off until an OTLP endpoint is set through the standard variables, and an allow-list of span attributes that keeps money, holdings, account numbers, tokens and query arguments out of every export (GR-12).

71. **Do plugins export their own spans?**
   - *Problem:* a plugin's spans would show what a connector does with the institution's answer. But the operator doesn't pass the OpenTelemetry variables to plugin pods, each plugin's NetworkPolicy lets it reach only the core and the egress proxy, and the egress proxy refuses private addresses, so a plugin can't reach a collector in the cluster.
   - *Today:* plugins built on the SDK trace every call with a no-op provider: the trace the core propagates continues through the plugin to its calls back to the core, so the core's spans stay connected, but the plugin records and sends nothing ([Plugins](concepts/plugins.md)).
   - *Options:* keep; or have the operator copy the core's OpenTelemetry settings into every plugin and add the collector to each plugin's NetworkPolicy, a new way out of a plugin pod that the grant doesn't name; or have plugins hand their spans to the core, which exports them, so plugins keep no way out but the proxy.
72. **Are calls to institutions traced?**
   - *Problem:* how long a broker or a rates API takes is the slowest part of a sync, but a span of such a call would name the host and the path, and paths such as EODHD's name the instruments the owner holds.
   - *Today:* calls through the egress proxy are not traced, and no trace context is ever sent outside the cluster; a sync's span covers the connector call as a whole.
   - *Options:* keep; or trace them with the method, the host and the status alone, never the path or the query.
73. **Metrics and logs over OTLP.**
   - *Problem:* the request was for tracing. OpenTelemetry also carries metrics and logs, which the instrumentation libraries in use (otelhttp, otelpgx, River's middleware) would record with little work.
   - *Today:* only traces are exported; the libraries' metrics go to OpenTelemetry's no-op meter, and logs stay JSON on stderr.
   - *Options:* keep; or export metrics with the same switch; or metrics and logs both, under the same privacy rule.
74. **Recording SQL statements.**
   - *Problem:* a query's span would be easier to read with its statement, and OpenFief's statements are parameterized, so the text holds no values; but a statement built with a literal would leak it, and nothing checks that none is.
   - *Today:* a query's span is named by its operation alone, such as `SELECT` or `INSERT`, and carries neither the statement nor its arguments.
   - *Options:* keep; or record the statement's text, never its arguments, after a check that no statement embeds a value.
