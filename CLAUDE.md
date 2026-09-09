@AGENTS.md

<!-- Add Claude Code-specific guidance here only when it has no AGENTS.md counterpart. -->

## Setup-Assistent (erster Lauf)

Wenn der Nutzer das Repo frisch geklont hat oder sagt "Setup", "erster Lauf", "onboarding": diesen Assistenten **in dieser Reihenfolge** durchspielen. Nicht überspringen, nichts erfinden. Ziel: User-Layer befüllen, Config aus Examples ableiten, alle UIs bauen und zeigen - ohne echte Bewerbung zu senden.

### A. Intake (nur Inhalte, Stil unangetastet)

Abfragen bzw. ablegen - jeweils nur, was der Nutzer liefert:

| Was | Wohin |
|-----|--------|
| Zeugnisse / Referenzen / Master-CV | `documents/` (siehe `documents/README.md`) |
| Bisherige Bewerbungen / Tracker | `data/applications.md` (Pipeline laut `DATA_CONTRACT.md` / `examples/`) |
| Anschreiben-Beispiele | `writing-samples/` (Layouts bleiben in `templates/`) |
| Interview-Infos (STAR, Verfügbarkeit, Gehalt, Standort/Remote) | `interview-prep/` + Fakten in `config/` aus den Examples |

**Harte Regel:** Nutzer-Inhalte befüllen `cv.md`, `config/profile.yml`, `portals.yml`, `documents/`, `data/`, `interview-prep/`, `writing-samples/` usw. - **`voice-dna.md` und Template-Layouts NIEMALS aus Nutzertext ableiten, überschreiben oder "anpassen".** `voice-dna.md` ist die feste Stil-Autorität dieses Starters (shared writing style). Layouts unter `templates/` (inkl. türkis) bleiben System-Layer.

### B. Config aus Examples

Aus den Vorlagen kopieren und mit Intake-Daten füllen - nichts erfinden, leere Felder leer lassen oder nachfragen:

1. `config/profile.example.yml` -> `config/profile.yml`
2. `config/cv-facts.example.json` -> `config/cv-facts.json`
3. `templates/portals.example.yml` -> `portals.yml`
4. `.env.example` -> `.env` (Keys nur, wenn der Nutzer welche hat)
5. Bei Bedarf: `modes/_profile.template.md` -> `modes/_profile.md`, `_custom.template.md` -> `_custom.md`, `_brief.template.md` -> `_brief.md`
6. Master-CV: aus `documents/` / Nutzertext -> `cv.md` (Keywords reformulieren, nie fabriizieren)

### C. Build aller UIs + Binaries, dann Tour

Toolchain prüfen und **bauen**, nicht nur erklären. Pro UI: Startbefehl, was man sieht, wofür, wann nutzen.

1. **Root (Healthcheck)**
   - `npm install`
   - `node doctor.mjs` (oder `npm run doctor`) - Checklist: Node, CLI, Playwright/Chromium, Config-Lücken. Fehlschläge erklären und beheben, bevor es weitergeht.
   - Optional PDF: `npx playwright install chromium`

2. **`web/` - Browser-Dashboard**
   - `cd web && npm install`
   - Dev: `npm run dev` (Next.js); Prod-Check: `npm run build`
   - Nutzen: Pipeline/Bewerbungen/PDF im Browser. Wann: Überblick, Tracker pflegen, ohne Terminal-TUI.

3. **`dashboard/` - Terminal-Pipeline/Stats**
   - Voraussetzung: Go 1.24+
   - `npm run build:dashboard` bzw. `npm run serve:dashboard`
   - Nutzen: TUI über Tracker (Filter, Sortierung, Report-Preview, Status). Wann: schnelle Terminal-Session, Stats ohne Browser.

4. **`gui/` - Desktop-GUI**
   - Voraussetzung: Go 1.24+ und C-Compiler (cgo/Fyne; Windows z. B. gcc aus MSYS2)
   - `npm run build:gui` -> erzeugt `gui/career-ops-gui` / `.exe` (Binary ist gitignored / nicht Teil dieses Starters)
   - Start: `npm run gui` oder `cd gui && go run . --path ..`
   - Nutzen: Pipeline + Analytics + Werkzeug-Katalog per Klick. Wann: Desktop-Bedienung statt CLI für die gleichen `.mjs`-Tools.

Kurz-Tour nach dem Build: Root-Doctor grün -> Web-Dev öffnen -> Dashboard oder GUI gegen dasselbe `data/` zeigen.

### D. Smoke-Test

Demo/Fixture durch Template/Pipeline jagen (z. B. `examples/` / `test-fixtures/`, `seed-fixture.mjs` falls vorhanden): Layout + `voice-dna`-Stil sichtbar machen. **Keine echte Bewerbung senden**, kein Portal submitten, keine Mail rausschicken - Draft/Preview only.
