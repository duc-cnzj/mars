---
title: 📋 Audit & Changelog
lang: en-US
---

# 📋 Audit & Changelog

mars records every key operation for audit tracking and session playback.

## Changelog

Records every deployment change of a project. `FindLastChangelogsByProjectID` returns the recent deployment history (with full Config / EnvValues), used for **rollback** to a historical version.

## Audit Events

Records key events such as login and deployment, attributed by **operator email** (`operator_email`):

| Role | Visibility |
|---|---|
| admin | all events |
| ordinary user | only their own events |

- Accessing others' events returns 404 (treated as non-existent), preventing audit-log ID enumeration
- Failed logins log a `[auth audit]` Warning

## Session Playback (screen recording)

- Container terminal sessions are **recorded** and can be replayed via the file service (`ShowRecords`)
- Ordinary users can replay only their own sessions; admins can replay all
- HTTP downloads go through `RequireFileAccess` (file owner / admin)

## Related APIs

| API | Description |
|---|---|
| `changelog.FindLastChangelogsByProjectID` | recent deployment records of a project |
| `event.List` / `event.Show` | audit event list / detail |
| `file.ShowRecords` | session recording playback |

> All audit operations are governed by the access-control model — see [Access Control](./access-control.md).
