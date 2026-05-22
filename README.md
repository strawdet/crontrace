# crontrace

A lightweight wrapper that records cron job execution history, durations, and exit codes to a local SQLite store.

---

## Installation

```bash
go install github.com/yourusername/crontrace@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/crontrace.git
cd crontrace && go build -o crontrace .
```

---

## Usage

Prefix any cron command with `crontrace run` to start recording executions.

**crontab example:**
```
*/5 * * * * crontrace run -- /usr/local/bin/backup.sh --incremental
0   2 * * * crontrace run --name "nightly-cleanup" -- /opt/scripts/cleanup.sh
```

**View execution history:**
```bash
crontrace list
crontrace list --name "nightly-cleanup"
```

**Inspect a specific job:**
```bash
crontrace show --last
crontrace show --id 42
```

**Sample output:**
```
ID   NAME              STARTED              DURATION   EXIT
42   nightly-cleanup   2024-05-10 02:00:01  3.24s      0
41   nightly-cleanup   2024-05-09 02:00:01  3.11s      0
40   nightly-cleanup   2024-05-08 02:00:02  —          1
```

By default, the SQLite database is stored at `~/.local/share/crontrace/history.db`. Override with the `CRONTRACE_DB` environment variable.

```bash
CRONTRACE_DB=/var/log/crontrace.db crontrace list
```

---

## License

MIT © [yourusername](https://github.com/yourusername)