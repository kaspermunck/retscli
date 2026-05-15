---
name: rets
description: >
  Query the official Danish legal information system (Retsinformation) for laws,
  executive orders, circulars, guidance, parliamentary bills, and their version
  history. Trigger when the user asks about Danish legislation, retsinformation.dk,
  specific paragraffer (§), lovbekendtgørelser, bekendtgørelser, cirkulærer,
  vejledninger, Folketinget lovforslag, or wants to look up a particular Danish
  act by name or by year/number. Examples: "what does § 7 of databeskyttelsesloven
  say", "show me the latest version of forvaltningsloven", "find all bekendtgørelser
  from Skatteministeriet in 2024", "what bills produced the new ældrelov", "history
  of amendments to GDPR-loven", "lov nr 502 af 23. maj 2018".
tools: [Bash]
---

# rets — Danish legal information skill

You help users query Retsinformation, Denmark's official legal information
system, using the `retscli` binary. Pick the right subcommand by triaging the
user's input, then chain a follow-up call when the first reveals an ID.

## Prerequisites

- `retscli` binary in PATH (built from `~/dev/retscli`).
- No authentication. Public API.
- **Rate limit:** 20 requests/hour, 50/day per IP. Don't loop over hundreds
  of laws — paginate sensibly or use one summarising call.

## When NOT to use this skill

- For Folketinget-side parliamentary data (individual MP votes, committee
  meetings, party membership, voting roll-calls), use the `ftcli` skill
  instead — it queries `oda.ft.dk` directly. `retscli bills` covers the
  light cross-reference (bill → enacted law), not the full parliamentary
  detail.
- For Danish company data (CVR, ownership, annual reports), use `virkcli`.
- For Danish statistics, use the `dst` skill.

## Entity triage — pick the right command

Classify the user's input first, then dispatch:

| User has / asks about | Command |
|---|---|
| Popular name ("databeskyttelsesloven", "forvaltningsloven") | `retscli resolve <name>` → then `get` |
| `<year>/<number>` or ELI URI (`/eli/lta/2018/502`) | `retscli get <year>/<number>` |
| Free-text topic / keyword | `retscli search "<query>"` |
| Specific paragraph (§) of a known law | `retscli get <id> --paragraph <n>` |
| "All bekendtgørelser from Skat in 2024" | `retscli list --type BEK --year 2024 --ressort skat` |
| "Most recent laws / orders" | `retscli recent --type LOV --limit 20` |
| "How has this law changed over time?" | `retscli history <id>` |
| Folketinget bill ("L 33", "lovforslag om klima") | `retscli bills <query>` or `retscli bills "L 33"` |
| Any endpoint not covered above | `retscli query /lovgivning/...` |

If the user names a popular law informally (e.g. "købeloven", "GDPR-loven"),
**resolve first** — it returns the canonical `(year, number)` plus a
confidence score. Then chain `retscli get` on the result.

## Commands

### `resolve` — popular name → canonical ID

```bash
retscli resolve databeskyttelsesloven
# → 2024/289  LBKH — LBK nr 289 af 08/03/2024  (Databeskyttelsesloven)  [confidence 1.00]
```

Use this every time the user gives a name, not a number. Empty result means
no match — fall back to `retscli search "<name>"`.

### `search` — full-text search

```bash
retscli search "<query>" [--type LOV] [--year YYYY] [--ressort "<ministry>"] [--sort <field>] [--limit N]
```

Searches the title field across all document types. Output columns: `<year>/<number>`,
type code, publication date, title.

Filters:
- `--type` is an exact code: `LOV`, `LOVH` (historical), `LBK`, `LBKH`,
  `BEK`, `CIR`, `VEJ`, `SKR`, `KEN`.
- `--ressort` is free-text matched server-side ("Justitsministeriet",
  "Skatteministeriet", or short forms).
- `--sort`: `signature_date | publication_date | effective_date | title | year | number`.
  Prefix with `-` for descending. Default order is ASC.

### `get` — fetch a document

```bash
retscli get 2018/502                       # full text + metadata
retscli get 2018/502 --paragraph 7         # extract one §
retscli get 2018/502 --markdown            # rendered Markdown export
retscli get 2018/502 --include full        # include case / actors / timeline
retscli get /eli/lta/2018/502              # ELI URI also accepted
```

Output shows law metadata (ID, type, ressort, dates, ELI URI, consolidated/
historical flag) followed by preamble + every chapter, paragraph group, and
paragraph in document order. Each § lists its stk (subsections) and litra
(letter clauses).

When the user asks about a specific section ("hvad siger § 7", "subsection 2
of paragraph 5"), always use `--paragraph` — it's faster, returns less text,
and avoids overwhelming the response.

### `list` — browse by type / year / ministry

```bash
retscli list --type BEK --year 2024 --ressort skat
retscli list --type LOV --year 2024 --sort -publication_date
```

Same endpoint as `search`, just without a free-text query. Requires at least
one filter.

### `recent` — newest documents

```bash
retscli recent --limit 20            # newest of any type
retscli recent --type LOV --limit 5
```

Sorted by `publication_date` descending. Use to answer "what's the latest..."
questions.

### `history` — consolidation/amendment chain

```bash
retscli history 2018/502
retscli history 2018/502 --paragraph 7    # changes affecting one §
```

Returns the full version chain of a law: original enactment, every amendment
(`entry_type=amendment`, document type `LOV Æ`), and every consolidation
(`entry_type=consolidation`, document type `LBK`). The latest entry with
`Valid to = (current)` is the in-force version.

Critical for "is this law still valid?" or "what's the current version?"
questions.

### `bills` — Folketinget lovforslag

```bash
retscli bills "klima"                # search bills by title
retscli bills "L 33"                 # fetch one bill
retscli bills --status Vedtaget --limit 10
retscli bills --search "skat" --enacted=true
```

Bills are the parliamentary track behind every enacted law. Each bill has an
FT number ("L 33", "B 12", "F 5"), a status (Stadfæstet, Forkastet,
3\. beh./Forkastet, etc.), and may link to an enacted law ID. Use this to find
the political background / forarbejder for a current law.

### `query` — raw passthrough

```bash
retscli query /lovgivning/2018/502/timeline
retscli query /lovgivning/cases/ --param search=klima --param limit=5
retscli query /lovgivning/2018/502/forarbejder
retscli query /lovgivning/keywords/ --param limit=10
```

Use for endpoints not covered by a first-class subcommand: timeline,
forarbejder (preparatory works), keyword index, actors, periods, version diff.
Returns pretty-printed JSON.

## Global flags

- `--raw` — raw upstream JSON body (debugging, custom parsing).
- `--json` — parsed struct, pretty-printed (scriptable, pipe to `jq`).

Pick one or neither (default is human-readable).

## Danish legal taxonomy (cheat sheet)

| Code | Danish | Meaning |
|---|---|---|
| LOV | Lov | Act of Parliament, original enactment |
| LOV Æ | Lov om ændring | Amendment to an existing law |
| LBK | Lovbekendtgørelse | Consolidated act (includes prior amendments) |
| BEK | Bekendtgørelse | Executive order (issued by a minister under a law's authority) |
| CIR | Cirkulære | Circular (intra-government instruction) |
| VEJ | Vejledning | Non-binding guidance |
| SKR | Skrivelse | Letter / notice |
| KEN | Kendelse | Decision / ruling |

`H` suffix on the code (e.g. LBKH, LOVH) marks a *historical* version
superseded by a newer one. To find the currently in-force consolidation for an
older `LOV`, run `retscli history <year>/<number>` and read the last entry.

**Section addressing:** Danish laws are split into paragraphs (`§ 1`,
`§ 2`, ...). Each paragraph has subsections (`Stk. 1`, `Stk. 2`, ...). Each
subsection may have letter clauses (`litra a`, `litra b`, ...) and number
clauses (`nr. 1`, `nr. 2`, ...). When the user references "paragraf 7, stk. 3,
litra b", use `--paragraph 7` and pick out the relevant `Stk. 3` → `litra b`
from the output.

## Common workflows

**1. "What does GDPR-loven say about consent?"**
```bash
retscli resolve databeskyttelsesloven       # → 2024/289 (LBK)
retscli get 2024/289 --paragraph 6          # § 6 covers samtykke
```

**2. "Show me the latest bekendtgørelser from Skat."**
```bash
retscli list --type BEK --ressort "Skatteministeriet" --sort -publication_date --limit 10
```

**3. "How has § 7 of databeskyttelsesloven changed since 2018?"**
```bash
retscli history 2018/502 --paragraph 7
```

**4. "What bills produced the new ældrelov?"**
```bash
retscli search "ældrelov" --type LOV --limit 3       # → 2024/1651
retscli query /lovgivning/2024/1651/cases            # bills that led to this law
retscli bills "L 17"                                  # full bill record (if that's the number)
```

**5. "Is forvaltningsloven still in force, and what's the current version?"**
```bash
retscli resolve forvaltningsloven           # → current consolidation
# or:
retscli history 2014/433                    # last entry with Valid to = (current) is in force
```

**6. "Find all laws Skatteministeriet has signed in 2024."**
```bash
retscli list --type LOV --ressort "Skatteministeriet" --year 2024 --limit 50
```

## Presenting results

- For a single document: lead with a one-line summary (popular title + current
  reference like "LBK nr 289 af 08/03/2024"), then show the relevant
  paragraph(s) verbatim. Don't dump a whole law unless asked.
- For lists: render the most important columns as a markdown table; note the
  total count. Truncate long titles.
- For history: present as a timeline; flag which entry is the current
  in-force version.
- For paragraphs: preserve `§`, `Stk.`, `litra` markers — they are how Danish
  lawyers cite the text.
- Always cite source: Retsinformation (retsinformation.dk).
- If the user is reasoning about current legal obligations, prefer the
  current consolidation (`LBK`) over the original `LOV` — and flag that you
  did so.

## Gotchas

- **Consolidated vs. original.** `retscli get 2018/502` returns the
  original 2018 enactment of databeskyttelsesloven (`historical=true`). The
  in-force version is the latest consolidation (`LBK`, e.g. 2024/289). When
  the user wants "the current law", resolve by name or run `history` first.
- **`--type` is exact-match.** A search for `--type LBK` will miss historical
  consolidations stored as `LBKH`. If you get zero results, try both, or
  drop the `--type` filter.
- **Sort default is ASC.** For "latest" questions, use `retscli recent` or
  pass `--sort -publication_date` explicitly.
- **Paragraph endpoint hits the base (original) version.** For the wording in
  the current consolidation, fetch the consolidated act with
  `retscli get <consolidated-year>/<number> --paragraph <n>`. (Resolve by
  name first to find the right ID.)
- **Rate limits are tight.** 20 req/hour, 50/day per IP. Don't fan out across
  many documents in one go — prefer a single `list` or `search` call with
  filters.
- **Bill numbers have a space.** Use `"L 33"` not `"L33"`. Quote the
  positional arg so the shell doesn't split it.
- **Resolve confidence < 1.0.** Treat results below ~0.7 as ambiguous and fall
  back to `search`.

## Error recovery

**"expected <year>/<number> or an ELI URI"** — the user gave you a name or
short form. Run `retscli resolve <name>` first; if that returns no match,
fall back to `retscli search "<name>"` and let the user pick.

**Empty `data` array** — no matching documents. Verify the type code,
broaden the year window, or drop the `--ressort` filter (the API matches it
literally — try the full ministry name rather than a slug).

**HTTP 429** — rate limit hit. Wait a few minutes; tell the user. Don't
retry in a loop.

**HTTP 404 on `get`** — the `(year, number)` pair doesn't exist for that
document type. Check via `retscli search` whether the law you want lives at a
different year (e.g. a later consolidation has its own year/number).

**Garbled Danish characters** — output is UTF-8; ensure the terminal locale
is set (`LC_ALL=en_US.UTF-8` or similar).

## Example walkthrough

**User:** "What does paragraf 7 stk. 2 of databeskyttelsesloven say, and is it
still in force?"

```bash
# 1. Resolve popular name → current consolidation ID.
retscli resolve databeskyttelsesloven
# → 2024/289  LBKH  LBK nr 289 af 08/03/2024  (Databeskyttelsesloven)  [1.00]

# 2. Fetch § 7 of the current consolidation.
retscli get 2024/289 --paragraph 7

# 3. Confirm with history — last entry should match 2024/289.
retscli history 2018/502 --paragraph 7
```

Respond with: the literal text of § 7 Stk. 2 (from step 2), then a one-line
note that the section is in force as of the 2024/289 consolidation (LBK nr
289 af 08/03/2024), and that the previous version came from § 7 of LOV nr
502 af 23/05/2018. Cite Retsinformation as the source.
