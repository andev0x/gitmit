# Gitmit Project Conventions

## Commit Messages
We follow a professional Conventional Commits structure with detailed bodies for complex changes.

### Structure
- **Subject:** Concise `type(scope): description` (aim for ~50 chars).
- **Body:** A blank line followed by a bulleted list of changes. Each bullet should be concise and professional.

### Example
```text
feat(formatter): implement line-length constraints and enhanced AI context

    - Add maxSubjectLength and maxBodyLength configuration options.
    - Update Formatter to support automatic line wrapping and subject-to-body
      overflow with blank line separation.
    - Enrich AI prompt with summarized git diff content for better commit
      body generation.
    - Refine wrapping logic to preserve multi-paragraph structures.
    - Add unit tests for the new formatting and wrapping behavior.
```

## Workflow
When suggesting commit messages, always analyze the full context using:
`git status && git diff HEAD && git log -n 5 --oneline`
