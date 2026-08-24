---
title: 🔐 Access Control
lang: en-US
---

# 🔐 Access Control

mars applies a **6-level access-control model** to every transport API (gRPC + HTTP). The authoritative list lives in [doc/access_control.md](https://github.com/duc-cnzj/mars/blob/master/doc/access_control.md) in the repository.

## The 6-level Model

| Level | Check | Description |
|---|---|---|
| 🆓 Public | matches the `biz.IsPublicMethod` allowlist | callable without any credentials |
| 🔑 Authenticated | gRPC login interceptor validates the Bearer token | any user with a valid token |
| 🛡️ Namespace-level | `RequireNamespaceAccessByName/ByID` / `RequireProjectAccess` | public namespaces for any logged-in user; private ones for admin / owner / members |
| 🏠 Owner-only | `RequireNamespaceOwner` | namespace creator only |
| ⭐ Admin-only | `RequireAdmin` | admin role only; exempt when the full method name exactly hits an allowlist entry |
| 📄 File owner/admin | `RequireFileAccess` | `file uploader == current user` or admin |

## The Three-layer Auth Chain

Every gRPC method passes through, in order:

1. **Login interceptor**: methods matching `IsPublicMethod` are passed through directly; the rest require a Bearer token, and the user is injected into the context after validation
2. **Authorize gate**: services implementing `Authorize` automatically call `RequireAdmin` (e.g. the file / repo services)
3. **In-method access control**: the method body starts by calling AccessBiz `Require*` / `Can*` methods (namespace / project / owner level)

## Common Scenarios

| Scenario | Permission |
|---|---|
| Login / settings / cluster info / pictures / version | 🆓 Public |
| Project management, deployment | 🛡️ Namespace-level (any private-namespace member) |
| Namespace Transfer / Delete / UpdatePrivate / SyncMembers | 🏠 Owner |
| File upload, repo management | 🔑 / ⭐ |
| Session playback, file download | 📄 owner or admin |
| Audit events | 🔑 ordinary users see only their own; admins see all |

## Security Audit Notes

- `namespace.UpdateDesc` currently has no owner/access check; tighten it if the description carries sensitive info (recommend going through the owner check)
- The deployment gate (Apply/WebApply) equals namespace accessibility — **any member** of a private namespace can deploy; to tighten to owner/admin, replace `RequireNamespaceAccessByID` with an owner check in `deploy/apply.go`
- Access tokens follow a "possession is authorization" model (`Revoke`/`Lease` do not check ownership), which is by design — do not add ownership checks that would break self-revocation
