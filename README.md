# md-tasks-notify

[![License MIT](https://img.shields.io/badge/license-MIT-royalblue.svg)](LICENSE)
[![Go version](https://img.shields.io/github/go-mod/go-version/powerman/md-tasks-notify?color=blue)](https://go.dev/)
[![Test](https://img.shields.io/github/actions/workflow/status/powerman/md-tasks-notify/test.yml?label=test)](https://github.com/powerman/md-tasks-notify/actions/workflows/test.yml)
[![Coverage Status](https://raw.githubusercontent.com/powerman/md-tasks-notify/gh-badges/coverage.svg)](https://github.com/powerman/md-tasks-notify/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/powerman/md-tasks-notify)](https://goreportcard.com/report/github.com/powerman/md-tasks-notify)
[![Release](https://img.shields.io/github/v/release/powerman/md-tasks-notify?color=blue)](https://github.com/powerman/md-tasks-notify/releases/latest)

![Linux | amd64 arm64 armv7 ppc64le s390x riscv64](https://img.shields.io/badge/Linux-amd64%20arm64%20armv7%20ppc64le%20s390x%20riscv64-royalblue)
![macOS | amd64 arm64](https://img.shields.io/badge/macOS-amd64%20arm64-royalblue)
![Windows | amd64 arm64](https://img.shields.io/badge/Windows-amd64%20arm64-royalblue)

A command-line tool to send daily notifications for actual tasks found in markdown files.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Supported Task Formats](#supported-task-formats)
- [Examples](#examples)
- [Configuration](#configuration)

## Features

- 📅 Filter tasks by date (today, yesterday, tomorrow, or custom ranges).
- 📧 Send notifications via email or output to stdout.
- 📝 Support for Obsidian Tasks emoji format.
- 🔍 Process multiple markdown files and whole directories.

## Installation

### From releases

Download the latest binary from the [releases page](https://github.com/powerman/md-tasks-notify/releases/latest).

### Using go install

```sh
go install github.com/powerman/md-tasks-notify@latest
```

## Usage

```
Usage of md-tasks-notify:
  -email string
        Send output to this email address instead of stdout
  -from-day int
        Start day relative to today (-1 for yesterday, 0 for today)
  -to-day int
        End day relative to today (1 for tomorrow)
```

### Basic Usage

Send daily notifications for tasks due today:

```sh
md-tasks-notify -email user@example.com path/to/*-tasks.md
```

Output to stdout (useful for testing or sending to another tool):

```sh
md-tasks-notify path/to/tasks_dir/
```

### Advanced Usage

Include past due tasks (yesterday) and future tasks (tomorrow):

```sh
md-tasks-notify -from-day -1 -to-day 1 -email user@example.com ~/notes/
```

Get tasks from yesterday only:

```sh
md-tasks-notify -from-day -1 -to-day -1 ~/notes/
```

Process tasks from stdin (useful to get output without file names):

```sh
cat *-tasks.md | md-tasks-notify -email user@example.com
```

### Cron Setup

Add to your crontab to receive daily notifications at 9 AM:

```cron
0 9 * * * md-tasks-notify -email your@email.com /path/to/notes/
```

## Supported Task Formats

This tool primarily supports the **Tasks Emoji Format** used by the Obsidian [Tasks plugin](https://publish.obsidian.md/tasks/Reference/Task+Formats/Tasks+Emoji+Format).

### Task Status Examples

```markdown
- [ ] Undone task
- [x] Completed task
- [/] In progress task
- [-] Cancelled task
```

### Task with Dates

```markdown
- [ ] Review documentation 📅 2024-01-15
- [ ] Call client ⏳ 2024-01-15
- [ ] Submit report 🛫 2024-01-10 📅 2024-01-15
```

#### Date Emoji

- 📅 [Due date](https://publish.obsidian.md/tasks/Getting+Started/Dates#Due+date)
- ⏳ [Scheduled date](https://publish.obsidian.md/tasks/Getting+Started/Dates#Scheduled+date)
- 🛫 [Start date](https://publish.obsidian.md/tasks/Getting+Started/Dates#Start+date)
- ... there are many more date emojis available, but this tool only handles the ones listed above.

## Examples

### Example Input

`project-tasks.md`:

```markdown
# Project Tasks

- [x] Setup project ✅ 2024-01-10
- [ ] Write documentation 📅 2024-01-15
- [ ] Review code ⏳ 2024-01-15
- [ ] Deploy to staging 📅 2024-01-20
- [-] Old feature ❌ 2024-01-05
```

`personal-tasks.md`:

```markdown
# Personal Tasks

- [ ] Buy groceries 📅 2024-01-15
- [ ] Call dentist ⏳ 2024-01-16
```

### Example Output

When run on 2024-01-15 with default settings:

```
project-tasks.md:
- [ ] Write documentation 📅 2024-01-15
- [ ] Review code ⏳ 2024-01-15

personal-tasks.md:
- [ ] Buy groceries 📅 2024-01-15
```

## Configuration

To send notifications via email, you need to configure SMTP settings through environment variables:

```sh
export SMTP_FROM="First Last <your-email@gmail.com>"
export SMTP_HOST=smtp.gmail.com
export SMTP_PORT=587
export SMTP_USERNAME=your-email@gmail.com
export SMTP_PASSWORD=your-app-password
```
