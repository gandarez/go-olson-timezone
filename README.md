# go-olson-timezone

[![Tests](https://img.shields.io/github/workflow/status/gandarez/go-olson-timezone/Tests/master?label=tests)](https://github.com/gandarez/go-olson-timezone/actions)
[![codecov](https://codecov.io/gh/gandarez/go-olson-timezone/branch/master/graph/badge.svg?token=ATAVS1K1CV)](https://codecov.io/gh/gandarez/go-olson-timezone)
[![Go Reference](https://pkg.go.dev/badge/github.com/gandarez/go-olson-timezone.svg)](https://pkg.go.dev/github.com/gandarez/go-olson-timezone)

A Golang library that tries to figure out your local timezone.

> This lib has been ported from [tzlocal](https://github.com/regebro/tzlocal) with some improvements for Go.

## Installation

```bash
go get -u github.com/gandarez/go-olson-timezone
```

## Features

* Always try to parse `TZ` environment variable if present.
* For Unix based systems (_including macOS_), timezone may be parsed from:
  * `/etc/timezone`
  * `/var/db/zoneinfo`
  * `/etc/sysconfig/clock`
  * `/etc/conf.d/clock`
  * `/etc/localtime`
* For windows, timezone will be discovered at registry key `SYSTEM\CurrentControlSet\Control\TimeZoneInformation`.
