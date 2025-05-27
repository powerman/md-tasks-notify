# md-tasks-notify

[![Go Reference](https://pkg.go.dev/badge/github.com/powerman/md-tasks-notify.svg)](https://pkg.go.dev/github.com/powerman/md-tasks-notify)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/powerman/md-tasks-notify/test.yml?logo=github&label=build)](https://github.com/powerman/md-tasks-notify/actions/workflows/test.yml)
[![Coverage Status](https://raw.githubusercontent.com/powerman/md-tasks-notify/gh-badges/coverage.svg)](https://github.com/powerman/md-tasks-notify/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/powerman/md-tasks-notify)](https://goreportcard.com/report/github.com/powerman/md-tasks-notify)
[![Release](https://img.shields.io/github/v/release/powerman/md-tasks-notify)](https://github.com/powerman/md-tasks-notify/releases/latest)

The tool to send daily notifications for actual tasks found in markdown files.

## Supported task formats

Initially this tool was designed to support "Tasks Emoji Format" used by Obsidian plugin
[Tasks](https://publish.obsidian.md/tasks/Introduction).
Support for other formats may be added later - open an issue if you need one.

## Usage

Use this command as a cron task to run daily:

```sh
md-tasks-notify -email user@example.com path/to/tasks/*.md
```

By default it'll send notification for "not done" tasks either scheduled or due today.
Optionally it can include tasks to be done in the near future and past due tasks.
