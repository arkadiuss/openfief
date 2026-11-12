---
title: The core's role can only add to the audit log; an UPDATE or DELETE fails even for a superuser
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:42.485Z'
id: SCEN-identity-audit-append-only
type: scenario
---
Given an identity database migrated by the core's role, when that role runs UPDATE, DELETE or TRUNCATE on identity.audit_log it is refused for want of privilege, and when a superuser does, the trigger raises; the entry is unchanged. Assumes the table is owned by the role that migrated, as the CloudNativePG app role is. Table owners could still drop the trigger, which this scenario does not cover.