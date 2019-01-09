# Advanced Cron
[![Build Status](https://travis-ci.org/UPDG/AdvancedCron.svg?branch=master)](https://travis-ci.org/UPDG/AdvancedCron)

Advanced cron scheduler written in Golang and compatible with Linux and Mac.

Advantages
* Compatible with legacy fron files
* Ability to create tasks for every second, not minute as default
* Ability to log output for each job individually
* Ability to choose different shell for each job
* Ability to set a timezone for the specific task and default for all
* Ability to set user for all jobs together
* Outputs execution time to stdout
* Errors can be logged into Sentry

## Configuration

You can use legacy cron configuration file (https://linuxconfig.org/linux-crontab-reference-guide)
or use new style config below.

By default daemon looking for config in the current directory.
To change that you can use -config and -path flags.

### New style config

Configuration is set via config file written in JSON, TOML, YAML, HCL, and Java properties.
By default daemon looking for config.* in the current directory.
To change that you can use -config (filename w/o extension) and -path flags.

Example of configuration in YAML: [config.example.yaml](config.example.yaml)

### Time values

#### Legacy format

Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | No         | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?


##### Special Characters

Asterisk ( * )

The asterisk indicates that the cron expression will match for all values of the
field; e.g., using an asterisk in the 5th field (month) would indicate every
month.

Slash ( / )

Slashes are used to describe increments of ranges. For example 3-59/15 in the
1st field (minutes) would indicate the 3rd minute of the hour and every 15
minutes thereafter. The form "*\/..." is equivalent to the form "first-last/...",
that is, an increment over the largest possible range of the field.  The form
"N/..." is accepted as meaning "N-MAX/...", that is, starting at N, use the
increment until the end of that specific range.  It does not wrap around.

Comma ( , )

Commas are used to separate items of a list. For example, using "MON,WED,FRI" in
the 5th field (day of week) would mean Mondays, Wednesdays and Fridays.

Hyphen ( - )

Hyphens are used to define ranges. For example, 9-17 would indicate every
hour between 9am and 5pm inclusive.

Question mark ( ? )

Question mark may be used instead of '*' for leaving either day-of-month or
day-of-week blank.


#### Human format

Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@reboot                | Run once after config load/reload          | 
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight on Sunday        | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *

#### Intervals

@every `duration`

For example, "@every 1h30m10s" would indicate a schedule that activates every
1 hour, 30 minutes, 10 seconds.

#### Run once

@once `time`

For example, "@once @every 1h" would indicate a schedule that job need to be 
run 1 time 1 hour after config load/reload.

You can set `time` in all formats described above, excluding `@reboot`.

Few examples:
```
@once 1 * * * *
@once @montly
@once @every 1d
```

#### Time zones

By default, all interpretation and scheduling is done in the machine's local
time zone (as provided by the Go time package http://www.golang.org/pkg/time).
The time zone may be overridden by providing an additional space-separated field
at the beginning of the cron spec, of the form "TZ=Asia/Tokyo"

Be aware that jobs scheduled during daylight-savings leap-ahead transitions will
not be run!

## Credits

Scheduler is based on https://github.com/robfig/cron (additional functionality added).