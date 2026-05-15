# retscli

A CLI for querying Danish legal documents via the Retsinformation REST API.

## Stack

- Go (1.25+), structured as a cobra CLI
- `rets/` package: Retsinformation API client + typed structs + table/CSV
  formatters
- `cmd/` package: cobra subcommands, one file per subcommand
- No auth required — the API is public

## CLI commands

```
retscli search <query>         # full-text search (title); filters: --type, --year, --ressort, --sort, --limit
retscli get <year>/<number>    # full law text; --paragraph <n>, --markdown, --include {case,actors,timeline,full}
retscli list --type LBKH       # browse by type / year / ministry (no free-text query)
retscli recent --type LOV      # latest published, sorted by publication_date desc
retscli history <year>/<number> # consolidation + amendment chain; --paragraph <n>
retscli resolve <name>          # popular name → (year, number) ID
retscli bills [query|L 33]      # Folketinget bills; positional doubles as bill number
retscli query <path>            # raw passthrough to any /v1/ endpoint (--param key=value)
```

Every data-returning command supports `--raw` (raw upstream JSON body) and
`--json` (parsed struct, pretty JSON). Default output is a human-readable table
or paragraph layout.

Accepts both `2018/502` and ELI URIs (`/eli/lta/2018/502`) as the document ID.

## API endpoint

`https://retsinformation-api.dk/v1` — a well-documented third-party REST mirror
with OpenAPI spec, version history, diff, paragraph addressing, and a join
against Folketinget bills.

Chosen over the official harvest service at `api.retsinformation.dk` (which
only exposes raw XML dumps) because retsinformation-api.dk gives us:

- Structured JSON for every document (paragraphs, stk, litra)
- Consolidation/amendment timeline across versions (`/history`, `/timeline`)
- Resolve endpoint for popular names → canonical IDs
- Markdown / PDF / diff export
- Bills (Folketinget) joined in under the same root

**Rate limits:** 20 requests/hour, 50/day per IP. Keep usage modest; cache
output for expensive walks.

## Document type codes (the `--type` flag)

- `LOV`  Lov (act, current)
- `LOVH` Lov, historical/superseded
- `LOVÆ` Lov, ændring (amendment) — note the Æ in the upstream `entry_type`
- `LBK`  Lovbekendtgørelse (consolidated act)
- `LBKH` Lovbekendtgørelse, historical
- `BEK`  Bekendtgørelse (executive order)
- `CIR`  Cirkulære
- `VEJ`  Vejledning (guidance)
- `SKR`  Skrivelse
- `KEN`  Kendelse

Filter is exact-match. To find a current consolidated act, you usually want
`LBKH` for laws old enough to have been re-consolidated; for the live version
follow up with `retscli history` to find the latest entry.

## Sorting

`--sort` accepts: `signature_date`, `publication_date`, `effective_date`,
`title`, `year`, `number`. Prefix with `-` for descending (e.g.
`--sort -publication_date`). The default API order is ASC; `retscli recent`
already sorts DESC for you.

## Build

```bash
go build -o retscli .            # builds binary at ./retscli
go install .                     # installs to $GOBIN (usually ~/go/bin)
```

`go.mod` pins `go 1.25.0` so the modern macOS toolchain emits `LC_UUID` —
older toolchains produced binaries dyld rejected with "missing LC_UUID load
command".

## Skill

A single skill `rets` ships in this repo at `.claude/skills/rets/SKILL.md` and
covers all subcommands — Claude triages the user's question to the right one.
See the README for install instructions.
