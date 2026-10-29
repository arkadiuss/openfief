---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Blocking real use

1. **XTB import needs the export's real layout.**
   - *Problem:* D-9 says XTB data arrives as uploaded xStation export files, but the export's columns, and whether it separates CFDs from stocks, are unverified (OQ-22).
   - *Today:* no XTB parser exists.
   - *Needed:* a header row and two or three rows with made-up values, copied from a real export. The parser's fixtures will be synthetic.
   - *Known so far:* the author holds stocks and ETFs only at XTB (so D-24's CFD handling stays as written) and will send the headers from a real export. The export's language is not known yet. Public parsers disagree on the sheet names (R-19).

## Plugins

2. **How the operator reads an in-cluster service in another namespace.**
   - *Problem:* D-74 has the operator add a NetworkPolicy rule to an in-cluster service's pods and hand the plugin its ClusterIP, which means reading that Service; D-15 confines the operator to a Role in `openfief`, and D-82 counts the cluster-wide permissions OpenFief holds.
   - *Today:* the owner's install (with the owner's kubeconfig) creates, in the service's namespace, a Role and RoleBinding `openfief-plugin-<instance>` that let the operator's ServiceAccount `get` that one Service; the operator reads it uncached, retries every minute when it can't, and doesn't watch it, so a changed selector or ClusterIP takes effect at the next reconcile. Uninstall deletes them.
   - *Options:* keep; have the install record the Service's ClusterIP, selector and target port in the grant so the operator reads nothing outside `openfief`; or a ClusterRole to read Services.

## Identity and the UI

3. **Two view plugins with the same browser path.**
   - *Problem:* D-81 has a view plugin's manifest declare the path prefix it serves to browsers and the grant approve it, but says nothing about two running instances that claim the same prefix, such as two installs of one view image.
   - *Today:* the core proxies a path to the instance with the longest matching browser path; when two instances claim that same prefix, it serves neither and answers 503, and logs both names. Nothing refuses the second install or grant.
   - *Options:* keep; have the CLI refuse an install whose browser path a running instance already serves; or pick one by a rule, such as the older instance.

4. **The core's own Connect services on the browsers' port.**
   - *Problem:* D-81 makes the core the only entry point for browsers, so `kubectl port-forward` and an owner's Ingress reach the core's port, where the AuthService and the owner's entry services (D-69, D-79) are served next to the read API. They answer only a view plugin's projected token, but they are reachable from outside, which REQ-identity-auth-not-public reads as "not offered".
   - *Today:* one port: the front door sends `/v1`, `/healthz`, `/readyz` and every `/openfief.*` path to the core, everything else to view plugins. The demo's Ingress routes every path to the core.
   - *Options:* keep; or serve the view-plugin-only services on a second port that the plugins' `connect` Service port targets and no browser route reaches.
