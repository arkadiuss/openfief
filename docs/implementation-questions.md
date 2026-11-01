---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Releases

54. **Whether a release ships the database.**
   - *Problem:* the install file runs the core against a CloudNativePG cluster `openfief-db` in `openfief` and reads its connection string from the Secret CloudNativePG makes for it (D-14), but docs/dev/sdlc.md names only the CRDs and the bootstrap install as release assets, and a `Cluster` manifest carries choices that depend on the owner's cluster: the storage class and size, the instance count, and the backups D-29 asks to be encrypted.
   - *Today:* the bundle ships neither CloudNativePG nor the cluster. [Install a release](guides/install.md#1-install-cloudnativepg-and-the-database) has the owner install CloudNativePG at the version the end-to-end tests run and apply a minimal one-instance `openfief-db` on the default storage class by hand; the end-to-end bundle run applies the demo's `deploy/demo/database` instead.
   - *Options:* keep the manual step; or ship a minimal `openfief-database-<version>.yaml` next to the bundle, the guide's cluster, for the owner to edit or replace; or ship it and CloudNativePG's release manifest pinned by sha256 as assets, and have the end-to-end run install from those.
