# Gitmit Algorithms

## Overview
Gitmit generates Conventional Commit messages by combining git diff parsing, heuristic analysis, weighted scoring, and template selection. The pipeline is fully offline and deterministic, with optional AI as a separate layer.

```
Git status/diff → Parser → Analyzer → Templater → Formatter → Commit message
```

## 1. Change Collection (Parser)
**Location:** `internal/parser/git.go`

1. **Staged file discovery:** `git status --porcelain` is scanned to identify staged files and their actions (A/M/D/R/C).
2. **Per-file diff extraction:** For each staged file, `git diff --cached -U0 -- <file>` is streamed.
3. **Line stats:** Added/removed lines are counted by diff prefixes (`+`/`-`).
4. **Major change flag:** A file is marked `IsMajor` when added+removed lines ≥ 500.

The parser returns a list of `Change` objects and aggregates totals for diff-stat analysis.

## 2. Analyzer: Feature & Context Extraction
**Location:** `internal/analyzer/analyzer.go`

### 2.1 File/Topic/Item Detection
- **Topic** is inferred from directory path with configurable overrides (`topicMappings`).
- **Item** defaults to the filename without extension.
- **Purpose** is inferred from keyword mappings and built-in keyword heuristics.

### 2.2 Symbol Extraction
Regex-based extraction detects structures from added lines:
- **Functions** (Go, JS/TS, Python, Java)
- **Structs/Classes**
- **Methods** (receiver-based Go methods)

These symbols are used to populate `{item}` placeholders and improve specificity.

### 2.3 Change Pattern Detection
Single-file patterns include:
- error handling, tests, imports, docs/comments, refactors
- API/database/performance/security indicators
- validation, logging, middleware, DI, CLI changes

### 2.4 Multi-file Pattern Detection
Across all changes, Gitmit detects patterns such as:
- **feature-addition** (many new files)
- **bug-fix-cascade** (many modified files with fix keywords)
- **refactor-sweep** (mixed A/M/D)
- **test-suite-update** / **config-update**
- **api-redesign** / **database-migration**

### 2.5 Special-Case Fallbacks
Early exits provide deterministic messages for clear cases:
- Single added file → `feat`
- Single deleted file → `chore`
- Only docs/config/deps → `docs`/`ci`/`chore(deps)`

## 3. Action (Type) Scoring Algorithm
The commit **action** is determined by a weighted score map, with support for normalized confidence weights (default).

### 3.1 Normalized Scoring (Default)
Gitmit uses **normalized confidence weights** to reduce noise when multiple signals compete.

1. **Normalize signals (0–1):**
   - **Branch hint:** 1.0 if branch name matches an action, 0.0 otherwise.
   - **Diff-stat:** 0–1 based on distance from thresholds (added/removed ratio).
   - **Keywords:** Raw keyword scores are normalized relative to the highest-scoring action.
   - **Multi-file patterns:** 1.0 if a relevant pattern is detected, 0.0 otherwise.
2. **Apply confidence weights:**
   - branch: 0.35
   - diff-stat: 0.25
   - keywords: 0.25
   - multi-file patterns: 0.15
3. **Final score:** `sum(weight × normalized_signal)` per action.
4. **Selection:** The action with the highest final score is selected.
5. **Fallback:** If top action score < 0.35, Gitmit falls back to file-based heuristics.

### 3.2 Legacy Additive Scoring
If `normalizeScoring` is disabled in config, Gitmit falls back to raw score aggregation:
1. **Branch name hints:** +3 to matching action.
2. **Diff-stat ratio:** +2 to `feat` or `refactor`.
3. **Keyword scoring:** per-action weights are added directly.
4. **Multi-file patterns:** +3 or +4 to relevant actions.

## 4. Scope Selection
- Single topic → that topic
- Single directory → directory name
- 2–3 topics → combined scope (sorted)
- Many topics → most common or `core`
- Commit history can override scope when consistent across recent commits

## 5. Template Selection & Scoring
**Location:** `internal/templater/templater.go`

1. **Template group resolution:** action → template group (A/M/D/R/DOC/SECURITY/MISC).
2. **Topic match:** exact → fuzzy → `_default`.
3. **Template scoring:**
   - Base score 1.0
   - +2.0 for matching detected patterns
   - +1.5 for using detected symbols
   - +1.0 for meaningful purpose placeholders
   - +0.5–1.5 for file-type relevance
   - +1.0 for major change templates
   - -0.5 for generic templates when specifics exist
4. **History de-dup:** recent messages are avoided when possible.

The highest-scoring template is selected, and placeholders (`{topic}`, `{item}`, `{purpose}`, `{source}`, `{target}`) are replaced.

## 6. Alternative Suggestions (Diversity Algorithm)
When regenerating suggestions:
- Used messages are filtered out.
- Similarity is computed using:
  - **Word-level Jaccard similarity (60%)**
  - **Character position matching (40%)**
- A diversity bonus favors less similar suggestions.
- A small random factor introduces controlled variation.

## 7. Configuration Influence
**Location:** `internal/config/config.go` + `docs/CONFIGURATION.md`

Configuration can adjust:
- Topic mappings
- Keyword mappings and weights
- Diff-stat thresholds
- Project-specific defaults

This allows the algorithm’s weighting to be tuned without code changes.
