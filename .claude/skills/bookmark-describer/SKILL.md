---
name: bookmark-describer
description: Use when the user asks to process untagged bookmarks, enrich bookmarks with descriptions, run the bookmark describer, describe bookmarks, tag unprocessed bookmarks, or automate bookmark metadata. This skill orchestrates end-to-end AI enrichment of linkding bookmarks that do not yet have the "mutated" tag.
---

# bookmark-describer — AI Bookmark Enrichment

Automates AI-generated descriptions and tags for linkding bookmarks that have not yet been processed (missing the `mutated` tag).

---

## 1. Pre-flight (always run first)

```bash
# 1. Verify connectivity — abort if this fails
linkdingctl config test

# 2. Fetch live tag list
linkdingctl tags --json | jq -r '.[].name' | sort

# 3. Fetch work queue (bookmarks without the mutated tag)
linkdingctl list --json --limit 0 | jq '[.results[] | select((.tag_names | contains(["mutated"]) | not))]'
```

Report the count of unprocessed bookmarks. If 0, stop and inform the user.

**Ask the user once (before processing any bookmark):**

> "Found N bookmarks to process. How should I proceed?
> - `preview` — show each proposal and prompt y/n/skip/all/quit before applying
> - `auto` — apply all without per-bookmark prompts (still prints what's being done)"

---

## 2. Per-bookmark loop

For each bookmark in the work queue:

### 2a. Print header
```
--- Bookmark #<id> ---
Title:  <title or website_title or domain>
URL:    <url>
Tags:   <tag_names>
```

### 2b. Fetch page content
Use `WebFetch` with a prompt to extract main content as plain text:
> "Extract the main content of this page as plain text. Include the title, headings, body text, and any key technical details. Exclude navigation, ads, footers, and boilerplate."

### 2c. Classify the fetch result

| Result | Condition | Action |
|--------|-----------|--------|
| **FULL** | >200 words of substantive content returned | Proceed with full analysis |
| **THIN** | <200 words but content is genuinely sparse (e.g. a stub page, short README, bare landing page) | Use `title` + `website_title` + `website_description` from JSON; append `[Limited content]`; still apply `mutated` |
| **BLOCKED** | 4xx HTTP error, SSL/TLS error, Cloudflare challenge page ("Just a moment..."), JS-only page with no readable body, or login/paywall wall with no metadata available | Skip entirely — do **NOT** add `mutated` — record in session summary as "blocked" |
| **FAILED** | Network error, timeout, unexpected exception | Skip entirely — do **NOT** add `mutated` — record in session summary as "failed" |

Key distinction: a 403, SSL error, Cloudflare challenge, or JS-heavy page with no readable body is classified **BLOCKED**, not THIN — even if the returned text is short. THIN is reserved for pages that loaded successfully but have little content.

Detection heuristics for BLOCKED:
- HTTP status 4xx or 5xx
- SSL/TLS certificate error in the fetch result
- Returned body contains "Just a moment" / Cloudflare challenge markers
- Returned body is <50 words and consists only of navigation/boilerplate (no meaningful sentences)
- `WebFetch` returns an error string rather than page content

Special cases:

| URL type | Handling |
|----------|----------|
| PDF | Process if WebFetch returns text; use THIN fallback if empty |
| GitHub repo | Focus on README content |
| YouTube video | Use title + description from bookmark JSON; note `[Video — limited text content]` |

### 2d. Write description

Apply the Description Writing Rules (section 4). Silently self-check before outputting.

### 2e. Select tags

Apply the Tag Recommendation Rules (section 5).

### 2f. Preview mode gate

- **`preview` mode:** Display proposal (description + tags), then prompt:
  > `Apply this update? [y/n/skip/all/quit]`
  - `y` — execute update
  - `n` — skip, record as rejected in summary
  - `skip` — skip silently
  - `all` — execute this and all remaining without further prompts (switches to auto mode)
  - `quit` — stop processing; print partial summary
- **`auto` mode:** Print proposal and proceed immediately.

### 2g. Execute update

```bash
linkdingctl update <id> --description "<text>" --add-tags "<tag1>,<tag2>,mutated"
```

**Always use `--add-tags`** (never `--tags`) to preserve existing user-applied tags.

On CLI error: print the error, mark as failed in session summary, do not retry.

### 2h. Optional spot-check (after update)

```bash
linkdingctl get <id> --json | jq '{id, description, tag_names}'
```

---

## 3. Session summary (print at end)

```
=== Session Summary ===
Updated:  N
Blocked (403/SSL/JS — not tagged):  N
Skipped (network error — not tagged):  N
Rejected (user declined): N

Untagged bookmarks (will remain in work queue):
  - ID <id>: <url> — <reason>
```

---

## 4. Description Writing Rules

- **2–4 sentences, ~50–100 words**
- **Lead with the core topic** — never open with "This article explores...", "This post covers...", or similar setup phrases
- **At least one concrete specific** — tool name, version, named pattern, specific gotcha, command, or configuration detail
- **Utility test:** someone reading it 6 months later understands what they saved and why
- **No filler:** ban "deep dive", "comprehensive guide", "everything you need to know", "in this post"

Silent self-check before outputting; rewrite if any check fails:
- [ ] Does not start with "This article/post/guide..."
- [ ] Contains at least one concrete specific
- [ ] 50–100 words
- [ ] Passes the 6-month utility test
- [ ] No filler phrases

---

## 5. Tag Recommendation Rules

- Use the **live tag list fetched in pre-flight** — do not use a static embedded list
- Pick **2–4 tags** from existing tags
- Only invent new tags if the topic is genuinely absent from the existing list
- New tag format: `lowercase_underscore` noun or noun-phrase (e.g. `container_security`, `go_modules`)
- Always add `mutated` as the final tag via `--add-tags`
- **Never use `--tags`** on `update` — it replaces all tags and destroys existing ones

---

## 6. Key Commands Reference

```bash
# Work queue
linkdingctl list --json --limit 0 | jq '[.results[] | select((.tag_names | contains(["mutated"]) | not))]'

# Work queue count only
linkdingctl list --json --limit 0 | jq '[.results[] | select((.tag_names | contains(["mutated"]) | not))] | length'

# Live tag list
linkdingctl tags --json | jq -r '.[].name' | sort

# Update bookmark (preserves existing tags)
linkdingctl update <id> --description "<text>" --add-tags "<tag1>,<tag2>,mutated"

# Spot-check result
linkdingctl get <id> --json | jq '{id, description, tag_names}'
```

---

## 7. Edge Cases

| Situation | Handling |
|-----------|----------|
| Bookmark already has non-empty description | Overwrite it — no `mutated` tag means not yet AI-enriched |
| Empty title and empty website_title | Use domain extracted from URL |
| `all` mode active | No per-bookmark prompts; execute and print as it goes |
| PDF URL | Process if WebFetch returns text; THIN fallback if empty |
| GitHub repo | Focus on README content in analysis |
| YouTube video | Use title + description from bookmark JSON; append `[Video — limited text content]` |
| Paywall / login wall — metadata available | THIN fallback; append `[Limited content — paywall or login required]`; apply `mutated` |
| Paywall / login wall — no metadata (403, zero body) | BLOCKED — do NOT apply `mutated` |
| Cloudflare challenge page ("Just a moment...") | BLOCKED — do NOT apply `mutated` |
| SSL/TLS certificate error | BLOCKED — do NOT apply `mutated` |
| JS-heavy page, no readable body (<50 words of boilerplate) | BLOCKED — do NOT apply `mutated` |
| CLI update fails | Print error, record in summary, move to next bookmark |
