---
title: Reference
description: Generated from the repository on every build, so these pages cannot go stale.
---

The [requirements report](/requirements/) is kibi's own view of every requirement with its evidence. It is strict on purpose: it counts a requirement as tested only when an end-to-end test of one of its scenarios has a fresh, passing receipt from `kibi prove`. OpenFief's unit and integration tests run on every commit but never count there, and no receipts are recorded yet, so the report shows none tested. The [status page](/status/) shows which requirements have code and tests behind them.
