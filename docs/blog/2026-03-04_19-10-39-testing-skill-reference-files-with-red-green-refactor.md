# Testing Skill Reference Files with Red-Green-Refactor

**Date:** March 4, 2026
**Author:** Rod Stewart (with Claude Opus 4.6)
**Project:** linkdingctl Claude Skill
**Task:** Adding advanced jq scripting recipes as a companion reference file

## The Challenge

The linkdingctl CLI has limited native filtering. `list --tags` uses AND logic only, there's no `--exclude-tags`, no date range filtering, and no cross-field boolean queries. Users who need these capabilities have to improvise jq pipelines.

The goal was to add a `references/jq-recipes.md` companion file to the linkdingctl skill, following the same pattern used by the excalidraw-diagram skill (which references `references/color-palette.md` and `references/element-templates.md`). The companion file would contain tested jq recipes that Claude reads on-demand rather than loading into every conversation.

Simple enough. But testing revealed two non-obvious problems that would have shipped broken if we hadn't followed a TDD-inspired approach.

## The Implementation

Two changes:

1. **SKILL.md** - A single reference line after the existing jq subsection, pointing to the companion file
2. **references/jq-recipes.md** - 311 lines of recipes across 6 categories, plus a JSON schema reference with field names verified against the Go source code at `internal/models/bookmark.go`

The recipes covered gaps that the native CLI can't fill:

| Category | Why native CLI can't do it |
| --- | --- |
| Negative Tag Filtering | `--tags` only supports AND inclusion |
| Tag Set Operations | No OR logic, no empty-tag detection |
| Date Filtering | No `--since`/`--before` flags |
| Multi-Field Queries | Can't combine filters with OR/AND |
| Aggregation & Reporting | No domain counting or summary stats |
| Batch Operations | CLI operates on one bookmark at a time |

## Testing: The jq Syntax Pass

The first test layer was straightforward: create a JSON fixture with 8 carefully designed bookmarks covering edge cases (untagged, archived, shared, stale, multi-tagged, single-tagged), then run representative recipes from each category.

17 recipes tested. 17 passed. Every jq expression parsed and produced correct output.

This was necessary but not sufficient. Correct jq means nothing if Claude never reads the file.

## Testing: The Retrieval Failure

The real learning started with the retrieval tests. We ran the same prompt in three fresh conversations:

> "Find bookmarks tagged k8s OR docker added this month that are still unread"

### Test 1: Passive Pointer, No Read

The first version of the reference line in SKILL.md read:

```
For advanced jq recipes (negative filtering, date ranges, multi-field queries,
batch operations), read `references/jq-recipes.md`.
```

**Result:** Claude ignored the reference entirely. It correctly identified that `--tags` uses AND logic, then improvised its own solution: two separate CLI calls piped through `jq -s` with `unique_by(.id)` deduplication. The solution worked but was less efficient (2 API calls instead of 1) and missed `--limit 0`.

**Root cause:** The pointer was passive. Claude saw a query it felt it could handle with general knowledge and never checked the reference. The phrase "For advanced jq recipes" didn't create enough urgency for what Claude perceived as a routine filtering question.

### Test 2: Strong Directive, Wrong Path

We strengthened the pointer:

```
**Important:** The native CLI cannot do OR-tag logic, negative tag filtering,
date range filtering, or cross-field boolean queries. Before improvising jq
pipelines for these, **read `references/jq-recipes.md`** -- it has tested,
efficient single-query recipes and the complete JSON field reference.
```

**Result:** Claude followed the directive. It explicitly quoted the "before improvising" instruction and attempted to read the file. But the path failed. Claude tried `Read(references/jq-recipes.md)` which resolved against the user's current working directory, not the skill directory. Then it tried `Search(pattern: "references/**/*")` which found nothing. It fell back to improvising.

**Root cause:** When a skill is loaded via the Skill tool, its content is injected into context. Relative paths like `references/jq-recipes.md` have no anchor; Claude doesn't know the skill lives at `~/.claude/skills/linkdingctl/`. The excalidraw skill uses the same relative paths, and it likely has the same latent bug.

### Test 3: Absolute Path, Full Success

The fix was one character change in substance: use the absolute path.

```
**Important:** The native CLI cannot do OR-tag logic, negative tag filtering,
date range filtering, or cross-field boolean queries. Before improvising jq
pipelines for these, **read `~/.claude/skills/linkdingctl/references/jq-recipes.md`**
-- it has tested, efficient single-query recipes and the complete JSON field reference.
```

**Result:** Claude read the file, found the OR-tag recipe, combined it with date filtering, and produced the exact single-query pattern from the recipes. It also included `--limit 0` and correctly noted that `--unread` pre-filters server-side.

## The Three Iterations Summarized

| Test | Pointer Style | Behavior | Issue |
|------|--------------|----------|-------|
| 1 | Passive, relative path | Ignored reference, improvised | No urgency to check |
| 2 | Directive, relative path | Tried to read, path failed | Relative path unresolvable |
| 3 | Directive, absolute path | Read file, used tested recipe | Working |

## Key Insights

### 1. Passive Pointers Don't Interrupt Improvisation

Claude is a capable improviser. If a reference pointer just says "for more info, see X," Claude will often skip it when it believes it can solve the problem directly. The pointer needs to be **prescriptive**: explicitly name the capabilities that require the reference file and say "before improvising."

The difference between "For advanced jq recipes, read X" and "The native CLI cannot do Y. Before improvising, read X" is the difference between a suggestion and a gate.

### 2. Relative Paths in Skills Are a Trap

This is the non-obvious finding. Skills are injected into context as text, divorced from their filesystem location. A relative path like `references/foo.md` resolves against wherever Claude happens to be working, not where the skill lives.

This is likely a latent bug in any skill that uses relative companion file references. The excalidraw skill references `references/color-palette.md` and `references/element-templates.md` with relative paths. Those will fail in the same way if the user's working directory isn't the skill directory itself.

**The fix is simple: use `~/.claude/skills/<skill-name>/references/foo.md`.** The `~` expands correctly regardless of working directory.

### 3. Testing Skills Requires Testing the Agent, Not Just the Content

Validating jq syntax was easy and necessary. But the real bugs were in **how Claude interacts with the skill**, not in the content itself:
- Would Claude read the file? (No, without urgency)
- Could Claude find the file? (No, with relative paths)
- Would Claude use what it found? (Yes, once it could read it)

A subagent test against a fixture only validates content correctness. Fresh-conversation tests validate the full retrieval-to-application pipeline. Both are required.

### 4. Red-Green-Refactor Works for Documentation

The TDD cycle applied directly:

- **RED:** First test showed Claude ignoring the reference (baseline failure)
- **GREEN:** Strengthened directive got Claude to try reading it
- **RED again:** Path resolution failure (new failure mode)
- **GREEN:** Absolute path fixed resolution
- **Verification:** Same prompt, correct output from tested recipe

Each iteration addressed a specific observed failure, not a hypothetical one.

## The Companion File Pattern

For anyone building Claude skills with companion reference files, here's the pattern that works:

```markdown
**Important:** [Name the specific gaps that require the reference].
Before improvising, **read `~/.claude/skills/<skill>/references/<file>.md`**
-- it has [value proposition: tested, efficient, complete].
```

Three elements that matter:
1. **Name the gaps explicitly** - Don't say "advanced usage." Say "OR-tag logic, negative filtering, date ranges."
2. **Say "before improvising"** - This interrupts Claude's tendency to solve things from general knowledge.
3. **Use absolute paths with `~`** - Relative paths resolve against the working directory, not the skill directory.

## What Shipped

- `SKILL.md`: 472 lines (2 lines added from original 470)
- `references/jq-recipes.md`: 311 lines across 6 categories
- All 13 JSON field names verified against `internal/models/bookmark.go`
- 17/17 jq recipes validated against test fixture
- Retrieval test: subagent found and applied recipes correctly
- Application test: fresh conversation produced correct compound query from recipes

The recipes file only loads when a user asks about filtering that the CLI can't do natively. Every other conversation pays zero context cost.

---

## References

- **Skill:** [SKILL.md](../../SKILL.md)
- **Recipes:** [jq-recipes.md](../../references/jq-recipes.md)
- **Bookmark model:** [bookmark.go](https://github.com/rodmhgl/linkdingctl/blob/main/internal/models/bookmark.go)
- **Pattern source:** [excalidraw-diagram skill](~/.claude/skills/excalidraw-diagram/SKILL.md) (companion file pattern)
