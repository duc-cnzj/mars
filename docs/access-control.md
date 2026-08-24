---
title: 🔐 Permissions
lang: en-US
---

# 🔐 Permissions

Mars keeps permissions simple: **namespaces** are the unit of access, and there are just a few roles.

## Roles

| Role | Who | What they can do |
|---|---|---|
| **Admin** | the account that runs the platform | everything, across all namespaces |
| **Owner** | the user who created a namespace | manage that namespace — transfer it, change public/private, sync members |
| **Member** | users added to a private namespace | deploy and manage apps inside it |
| **Logged-in user** | anyone with a valid account | use public namespaces and their own files |

## Public vs private namespaces

- **Public** — anyone logged in can see it and deploy into it. Good for shared staging/demo spaces.
- **Private** — only the owner and the members they add can access it. Good for production or sensitive environments.

> For privacy, a private namespace appears as if it doesn't exist to people without access.

## What each action needs

| Action | Who can do it |
|---|---|
| Deploy / upgrade an app | anyone with access to the namespace (members of private ones too) |
| Create a namespace | any logged-in user |
| Make a namespace private / transfer / delete / sync members | only the **owner** (or admin) |
| Open terminals / read logs | anyone with access to the namespace |
| Download your session recordings | the file's owner (or admin) |
| View all audit events | admin only — regular users see their own |

## Manage your team

To give someone access to a private namespace:

1. Open the namespace.
2. Use **sync members** to add the users.
3. Save — they can now deploy and manage apps in it.

To hand a namespace to someone else, use **transfer** — ownership moves to the other user.

## Notes for admins

- The built-in `admin` account bypasses all namespace-level checks.
- Two things to be aware of when you deploy Mars:
  - **Anyone with namespace access can deploy** into it (including private namespaces). If you need stricter controls for high-risk environments, you may want to gate deployments to the owner or admins.
  - **Whoever has an access token can use it** (tokens are treated as a secret you hold = the right to use it).

Need to set up login for your company? See [Configuration → Sign-in](./configuration.md#sign-in).
