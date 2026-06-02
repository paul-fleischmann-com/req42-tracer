# Project Roles

Dieses Dokument definiert die Rollen und Verantwortlichkeiten im req42-tracer Projekt.

## Rollen

### Developer — `dev-paul-fleischmann`

**Verantwortung:** Git-Identität für Commits

- Feature-Branches erstellen und implementieren
- Commits signieren (`git config user.name "dev-paul-fleischmann"`)
- Tests schreiben (gemäß [`TESTS.md`](TESTS.md))
- CI-Fehler beheben

> **Hinweis:** `dev-paul-fleischmann` wird ausschließlich als Git-Identität verwendet.
> Alle `gh` CLI Aufrufe (PRs öffnen, Issues erstellen, API-Calls) laufen über **`paulefl`**,
> da `dev-paul-fleischmann` ein sehr niedriges API-Rate-Limit hat (60 req/h).

**Branch-Konvention:**
```bash
git checkout -b feat/<issue-#>-kurzer-name
git config user.name "dev-paul-fleischmann"
git config user.email "dev@paul-fleischmann.com"
```

**PR erstellen (als paulefl):**
```bash
gh pr create --assignee dev-paul-fleischmann --reviewer paulefl
```

---

### Reviewer — `paulefl`

**Verantwortung:** Alle `gh` CLI Operationen, Code-Review, Merge

- Einziger Account für `gh` CLI Aufrufe (Issues, PRs, API, Milestones)
- Pull Requests reviewen (Code Review + Security Review gemäß [`REVIEW.md`](REVIEW.md))
- Review-Findings als Inline-Kommentare im PR dokumentieren
- PRs approven und in `master` mergen
- Releases taggen

---

## Workflow

```
git identity:                    gh CLI:
dev-paul-fleischmann             paulefl
        │                           │
        │  feature branch            │
        ├── git commit/push ──>      │
        │                            │
        │                            │  gh pr create
        │                            ├──────────────>
        │                            │  code review
        │                            │  inline comments
        │  fix findings  <───────────┤
        │  git commit/push ──>       │
        │                            │  approve + merge
        │  <─────────────────────────┤
```

## Account-Konfiguration

```bash
# Git-Identität für Commits (pre-commit hook prüft dies)
git config user.name "dev-paul-fleischmann"
git config user.email "dev@paul-fleischmann.com"

# gh CLI — immer paulefl (einziger aktiver Account)
gh auth switch --user paulefl

# Status prüfen
gh auth status
git config user.name
```

**Regel:** `gh auth switch` wird nicht mehr auf `dev-paul-fleischmann` gesetzt.
`paulefl` ist dauerhaft der aktive gh-Account.

## GitHub Konfiguration

| Setting | Wert |
|---|---|
| Default branch | `master` |
| Branch protection | PR required, 1 approval (`paulefl`) |
| Git commits | `dev-paul-fleischmann` (Identität in git config) |
| gh CLI / API | `paulefl` (einziger aktiver Account) |
| Review & Merge | `paulefl` (admin) |
</content>
