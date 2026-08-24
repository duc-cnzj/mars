---
title: 🐳 Containers & Logs
lang: en-US
---

# 🐳 Containers & Logs

Mars lets you work with your running containers straight from the browser — no `kubectl` needed.

## Open a terminal

Click a container and choose **terminal**. A shell opens inside that container so you can debug live — check config files, run commands, poke around.

- A **one-off command** option runs a single command and shows you the output.
- Need to replay what happened earlier? Terminal sessions are recorded — see [Audit & History](./audit.md).

## Watch live logs

Pick a container and open its **logs**. Output streams in real time, so you can watch an app boot up or trace an error.

## Copy files

Move files between your machine and a container:

- **Into the container** — upload a file (e.g. a config or a hotfix) to a path inside the container.
- **Out of the container** — download a file from inside the container to your computer.

Useful for grabbing logs, export files, or dropping in a quick fix without rebuilding.

## Monitor resource usage

| View | What you see |
|---|---|
| Per container | CPU and memory usage of that container |
| Project / namespace | overall CPU and memory across the project or namespace |
| Pod ranking | which pods are using the most resources — handy for spotting a runaway |

## Permissions

To open terminals or read logs in a namespace, you need **access to that namespace** — public namespaces are open to anyone logged in; private ones require membership. See [Permissions](./access-control.md).
