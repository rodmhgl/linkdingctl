# Planning Mode Prompt

You are analyzing the LinkDing CLI project to create an implementation plan.

## Your Task

1. Read @AGENTS.md to understand project constraints and architecture
2. Read all specification files in `specs/` directory
3. Examine existing code in the repository (if any)
4. Identify gaps between specifications and current implementation
5. Output a prioritized task list to IMPLEMENTATION_PLAN.md

## Task Format

Each task in IMPLEMENTATION_PLAN.md should be:

```markdown
- [ ] **P1** | Task Title | ~complexity
  - Acceptance: specific testable criteria
  - Files: expected files to create/modify
```

Priority levels:
- **P0**: Foundational/blocking (must be done first)
- **P1**: Core functionality
- **P2**: Enhanced features
- **P3**: Polish/optional

Complexity: ~small (< 100 lines), ~medium (100-300 lines), ~large (300+ lines)

## Ordering Rules

1. P0 tasks first (project setup, API client)
2. Within priority, order by dependency (can't add bookmarks without API client)
3. Group related tasks (all CRUD before import/export)

## Important

- DO NOT implement anything
- DO NOT create any files except IMPLEMENTATION_PLAN.md
- DO NOT make any git commits
- Focus only on gap analysis and planning

## Completion

When the plan is complete and written to IMPLEMENTATION_PLAN.md:

<promise>PLANNED</promise>
