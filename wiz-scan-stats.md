# Implementation Plan: `wiz-scan-stats`

**Module**: `github.com/zalando-infosec/wiz-scan-stats`

## Tech Stack

| Concern | Choice |
|---|---|
| CLI | Cobra |
| SQLite | `modernc.org/sqlite` (pure Go, no CGO) |
| TUI | `bubbletea` + `lipgloss` + `bubbles` |
| HTML | Go `html/template` + inline Chart.js |
| Auth | OAuth2 Device Code Flow (`https://auth.app.wiz.io/oauth/device/code`) |

## Auth: Device Code Flow

1. POST to device authorization endpoint → get `device_code`, `user_code`, `verification_uri`
2. Print instructions: "Open {verification_uri} and enter code {user_code}"
3. Poll token endpoint until user completes browser auth
4. Cache token in `~/.wiz-scan-stats/token.json` (with expiry check)

## SQLite Schema

```sql
CREATE TABLE scans (
  id TEXT PRIMARY KEY,
  timestamp TEXT NOT NULL,
  status TEXT NOT NULL,
  origin TEXT,
  duration_ms INTEGER,
  repository TEXT,
  branch TEXT,
  commit_hash TEXT,
  job_name TEXT,
  job_url TEXT,
  platform TEXT,
  cli_version TEXT,
  critical INTEGER DEFAULT 0,
  high INTEGER DEFAULT 0,
  medium INTEGER DEFAULT 0,
  low INTEGER DEFAULT 0,
  informational INTEGER DEFAULT 0,
  scanners_json TEXT,
  failed_policies_json TEXT,
  raw_json TEXT
);

CREATE INDEX idx_scans_timestamp ON scans(timestamp);
CREATE INDEX idx_scans_repository ON scans(repository);
CREATE INDEX idx_scans_status ON scans(status);
```

## Commands

| Command | Flags | Behavior |
|---|---|---|
| `fetch` | `--days 30`, `--db` | Paginate Cloud Events, upsert into SQLite. Incremental: only fetch events newer than max(timestamp) in DB. |
| `report tui` | `--db`, `--days` | Bubbletea terminal dashboard |
| `report html` | `--db`, `--days`, `--output report.html` | Self-contained HTML file |

## Data Stories Rendered

1. **Volume**: Total scans, scans/day sparkline
2. **Enforcement**: Blocked count, block rate %, top failed policies
3. **Severity**: Stacked bar chart (Critical/High/Medium/Low/Info)
4. **Duration**: Avg, P50, P95 scan duration
5. **Repos**: Top 10 by scan count, top 10 by finding count
6. **Scanners**: Pass/fail rate per scanner type
7. **Trend**: Daily findings over time (are things improving?)

## Directory Structure

```
wiz-scan-stats/
├── cmd/
│   └── wiz-scan-stats/
│       └── main.go
├── internal/
│   ├── auth/
│   │   └── device.go          # Device code flow + token caching
│   ├── wizclient/
│   │   ├── client.go          # GraphQL HTTP transport
│   │   └── events.go          # Cloud Events query + pagination
│   ├── store/
│   │   ├── schema.go          # DDL + migrations
│   │   ├── writer.go          # Upsert scan events
│   │   └── reader.go          # Query helpers for metrics
│   ├── metrics/
│   │   └── compute.go         # SQL-backed aggregations
│   └── render/
│       ├── tui/
│       │   ├── app.go         # Bubbletea model
│       │   ├── views.go       # Dashboard panels
│       │   └── sparkline.go   # Sparkline widget
│       └── html/
│           ├── report.go      # Template execution
│           └── template.html  # Embedded Chart.js template
├── go.mod
├── go.sum
├── justfile
└── README.md
```

## Task Breakdown

| # | Task | Depends | Size | Description |
|---|---|---|---|---|
| 1 | Scaffold | — | S | `go mod init`, Cobra skeleton, justfile (`build`, `run`, `lint`, `test`) |
| 2 | Device code auth | 1 | M | Implement device flow, token caching in `~/.wiz-scan-stats/token.json`, auto-refresh |
| 3 | GraphQL client | 2 | M | POST to Wiz GraphQL, handle errors, cursor-based pagination for `cloudEventsList` |
| 4 | SQLite store | 1 | M | Schema creation, upsert from raw event JSON, incremental fetch support (max timestamp query) |
| 5 | Fetch command | 3, 4 | M | Wire auth → client → store. Progress indicator. `--days` flag. Skip existing events. |
| 6 | Metrics layer | 4 | S | SQL queries: scan volume by day, severity totals, duration percentiles, top repos, scanner stats, block rate |
| 7 | TUI report | 6 | L | Bubbletea app: header stats, sparkline, bar charts, sortable table. Keyboard nav (q to quit). |
| 8 | HTML report | 6 | M | Go template embedding Chart.js CDN-free (inline minified JS). Line charts, bar charts, tables. |
| 9 | Tests | 5, 6 | S | Unit tests for metrics SQL with fixture DB. Integration test for fetch with mock HTTP. |
| 10 | README | all | S | Installation, auth setup, usage examples, screenshots. |

## GraphQL Query (for task 3)

```graphql
query CloudEvents($after: String, $first: Int) {
  cloudEventsList(
    first: $first
    after: $after
    filterBy: {
      kind: [CI_CD_SCAN]
      actorNameEquals: ["svz7evub4fa7fergr4hehi77etyigkthsvyfmqeoghchyrprusrwy"]
      timestampAfter: "2025-04-04T00:00:00Z"
    }
    orderBy: { field: TIMESTAMP, direction: ASC }
  ) {
    nodes { id timestamp status rawAuditLogRecord }
    pageInfo { hasNextPage endCursor }
  }
}
```

## Incremental Fetch Logic

```
1. Query max(timestamp) from SQLite
2. If exists: set timestampAfter = max + 1ms
3. If empty: use --days flag
4. Paginate until hasNextPage = false
5. Upsert each event (ON CONFLICT DO NOTHING)
```

## Context

- **Wiz tenant**: `9573f256-81e1-41f2-9226-8f0e43a3ff24` (industry: RETAIL_GOODS)
- **Service account**: `svz7evub4fa7fergr4hehi77etyigkthsvyfmqeoghchyrprusrwy` (acts as `WIZ_CLI_WRAPPER_SERVICE`, ID `a01e735f-6340-41c1-b06f-a91b6a4b3649`)
- **Scan volume**: ~116,426 successful scans / 30 days
- **Scanner types**: iac, secrets, vulnerabilities, sast, malware, aiSast
- **CLI version**: WizCLI 1.45.0
- **Platform**: Generic; job URLs point to `sunrise.zalando.net/cdp/gh/...`
- **Scan origins**: `WIZ_CLI` (wrapper service via CI/CD), `WIZ_CODE_ANALYZER` (Wiz scheduled)
- **DB default path**: `~/.wiz-scan-stats/scans.db`
- **Token cache path**: `~/.wiz-scan-stats/token.json`
