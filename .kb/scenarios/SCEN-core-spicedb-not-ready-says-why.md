---
title: A core under SpiceDB that is not ready says what is missing, by name
status: active
origin:
  kind: agent
  recorded_at: '2026-11-22T08:40:35.290Z'
id: SCEN-core-spicedb-not-ready-says-why
type: scenario
---
Given a core running --authz=spicedb, the default, when the file of SpiceDB's preshared key is missing, then /readyz answers 503 naming the Secret openfief-spicedb-key, the file and the credentials Job that makes it, and the log warns the same at the first look and every minute; when the key is there but the start's reconcile fails, then the core asks its database's catalog for SpiceDB's role and database spicedb and /readyz and each failed attempt's warning say that the openfief-db cluster has no role spicedb, no database spicedb or neither, and that the release's database file adds them (install step 5), or that SpiceDB refuses the core's key, or that SpiceDB at its address has not answered; every authorization check stays refused meanwhile, and once SpiceDB answers the core is ready. Assumes every PostgreSQL role may read pg_roles and pg_database, and that SpiceDB's role and database live in the cluster the core's own database is in.