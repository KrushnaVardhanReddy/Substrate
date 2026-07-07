# Substrate Prompts

This directory contains all Jules and Stitch task prompts, organized by development phase.

## Structure

```
prompts/
├── phase-0-specs/        # Spec writing tasks (handled by Antigravity, rarely Jules)
├── phase-1-diff-engine/  # Go CLI diff engine tasks
│   ├── t01_go_scaffold.txt
│   ├── t02_openapi_parser.txt
│   ├── t03_diff_comparator.txt
│   ├── t04_rule_engine.txt
│   ├── t05_cli_output.txt
│   └── t06_unit_tests.txt
├── phase-2-github-app/   # TypeScript GitHub App tasks
└── phase-3-dashboard/    # Go API + SvelteKit dashboard tasks
```

## How to Submit a Task to Jules

```bash
# List all available tasks
python3 jules_submit.py --list

# Submit a specific task by number
python3 jules_submit.py --task 1

# Submit a custom prompt file directly
python3 jules_submit.py --file prompts/phase-1-diff-engine/t01_go_scaffold.txt

# Target a specific branch
python3 jules_submit.py --task 1 --branch feat/diff-engine
```

## Prompt File Convention

Each prompt file follows this structure:

```
MANDATORY RULES — VIOLATION = REJECTED PR:
[safety rules prepended automatically by jules_submit.py]

---

TASK: [ID] — [Name]

═══════════════════════════════════════════════════════════════
OBJECTIVE
═══════════════════════════════════════════════════════════════
[What Jules needs to accomplish]

═══════════════════════════════════════════════════════════════
DELIVERABLES
═══════════════════════════════════════════════════════════════
[Exact files to create/modify with full specifications]

═══════════════════════════════════════════════════════════════
FILES LIST
═══════════════════════════════════════════════════════════════
FILES TO ADD:
- path/to/file

FILES TO MODIFY:
- path/to/file

Commit: "jules: [description]"
```

## Spec-First Rule

**Jules prompts must reference `docs/specs/` files as their source of truth.**
If a task requires a spec that doesn't exist yet, write the spec first (with Antigravity), then write the Jules prompt.
