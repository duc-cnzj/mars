---
title: 📋 Audit & History
lang: en-US
---

# 📋 Audit & History

Mars keeps a record of what happens on your platform — so you always know who did what, and can roll back when needed.

## Deployment history

Every time you deploy or upgrade an app, Mars saves it as a version with the full configuration. This gives you two things:

- **See what changed** — open a project's history to compare deployments.
- **Roll back** — revert to any previous version with one click.

See [Deploying Apps → Roll back](./projects.md#roll-back) for the steps.

## Operation records

Key events — logins, deployments and similar — are recorded with the person who did them.

| Who | What they can see |
|---|---|
| Admin | every event across the platform |
| Regular user | only their own events |

## Session recordings

Container terminal sessions are **recorded** while you use them. You can replay a session later:

- You can always replay **your own** sessions.
- Admins can replay **anyone's** sessions.

This is useful for security audits and for figuring out what exactly happened during an incident.

## Permissions

Everything here follows the [permission model](./access-control.md): downloading a session recording requires that you own it (or are an admin), and viewing others' events is limited to admins.
