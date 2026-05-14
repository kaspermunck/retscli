# retscli

A command-line client for the Retsinformation REST API — Denmark's official
legal information system (laws, executive orders, circulars, guidance,
parliamentary bills).

Part of a suite alongside [`dstcli`](https://github.com/kaspermunck/dstcli)
(Statistics Denmark) and [`virkcli`](https://github.com/kaspermunck/virkcli)
(Danish company registry).

## Install

```bash
go install github.com/kaspermunck/retscli@latest
```

Or build from source:

```bash
git clone <repo>
cd retscli
go build -o retscli .
```

## Usage

```bash
retscli search "databeskyttelse" --limit 5
retscli get 2018/502 --paragraph 7
retscli history 2018/502
retscli resolve databeskyttelsesloven
retscli recent --type BEK --limit 10
retscli bills "L 33"
retscli query /lovgivning/2018/502/timeline
```

Every data-returning subcommand supports `--raw` (upstream JSON body) and
`--json` (parsed struct, pretty JSON). Default output is a human-readable
table or paragraph layout.

## Commands

| Command | Use it for |
|---|---|
| `search <query>` | Free-text search across all documents (filterable by `--type`, `--year`, `--ressort`) |
| `get <year>/<number>` | Fetch a specific document; `--paragraph` to extract one section |
| `list` | Browse by `--type` / `--year` / `--ressort` (no free-text query) |
| `recent` | Most recently published, newest first |
| `history` | Consolidation + amendment chain for a law |
| `resolve` | Popular law name → canonical (year, number) |
| `bills` | Folketinget bills (lovforslag) |
| `query` | Raw passthrough to any `/v1/` endpoint |

Run `retscli <command> --help` for the full flag set.

## Document type codes

- `LOV` / `LOVH` — Lov (act) / historical version
- `LBK` / `LBKH` — Lovbekendtgørelse (consolidated act) / historical
- `BEK` — Bekendtgørelse (executive order)
- `CIR` — Cirkulære
- `VEJ` — Vejledning (guidance)
- `SKR` — Skrivelse
- `KEN` — Kendelse

Filter is exact-match; the consolidated current version of an older act is
typically stored as `LBKH` (the older `LOV` is then `historical=true`).

## API

This CLI talks to `https://retsinformation-api.dk/v1` — a public REST mirror
with full OpenAPI docs at `/docs`. No authentication required.

**Rate limit:** 20 requests/hour, 50/day per IP.

## Repository layout

```
.
├── main.go              # thin entry point
├── cmd/                 # one file per cobra subcommand
│   ├── root.go
│   ├── search.go
│   ├── get.go
│   ├── list.go
│   ├── recent.go
│   ├── history.go
│   ├── resolve.go
│   ├── bills.go
│   ├── query.go
│   └── helpers.go
└── rets/                # API client package
    ├── client.go        # HTTP client + APIError
    ├── types.go         # response structs
    └── format.go        # table/CSV printers
```

## License

MIT (or your preference — repo currently has no LICENSE file).
