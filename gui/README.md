# gui/ — grafische Oberfläche für career-ops

Ein eigenständiges Fenster-Programm über derselben Pipeline, die das Terminal-
Dashboard unter `dashboard/` anzeigt — plus einem Werkzeugbereich, über den die
Kommandozeilen-Skripte des Repos per Klick laufen.

## Warum es das neben `dashboard/` gibt

`dashboard/` ist ein TUI und setzt Terminal-Bedienung voraus (Tastenkürzel wie
`f`, `s`, `v`, `D`). Diese Oberfläche zeigt dieselben Daten, macht aber jede
Aktion sichtbar anklickbar und stellt zusätzlich die `.mjs`-Werkzeuge bereit,
die man sonst von Hand aufrufen müsste. Das TUI bleibt unverändert bestehen.

## Starten

```bash
npm run gui                # aus dem Repo-Wurzelverzeichnis
npm run build:gui          # baut gui/career-ops-gui.exe
```

Direkt:

```bash
cd gui
go run . --path ..                 # gegen das übergeordnete Repo
./career-ops-gui.exe               # Verzeichnis wird automatisch gesucht
./career-ops-gui.exe --path C:\pfad\zu\career-ops
```

Ohne `--path` sucht das Programm das career-ops-Verzeichnis: erst das
Arbeitsverzeichnis, dann bis zu drei Ebenen oberhalb der Programmdatei. Erkannt
wird es an `AGENTS.md`, `doctor.mjs` oder `tracker.mjs`.

## Voraussetzungen

- **Go 1.24+** und ein C-Compiler zum Bauen (Fyne benötigt cgo; unter Windows
  z. B. gcc aus MSYS2, unter Linux die üblichen X11/GL-Entwicklungspakete).
- **Node.js** zur Laufzeit — aber nur für den Werkzeugbereich. Ohne Node lassen
  sich Bewerbungen und Auswertungen ansehen; der Start eines Werkzeugs meldet
  dann verständlich, dass Node fehlt.
- Einmalig `npm install` im Repo-Wurzelverzeichnis, sonst brechen die Werkzeuge
  mit `Cannot find package 'dotenv'` ab.

## Die drei Bereiche

**Bewerbungen** — die Liste aus `data/applications.md` mit Freitextsuche,
Statusfilter und Sortierung (Nummer, Bewertung, Datum, Status, Firma), rechts
die Detailansicht: Kopfdaten, aus den Notizen abgeleitete Felder (Ort,
Arbeitsmodell, Gehaltsrahmen, letzter Kontakt), Statuspflege mit optionaler
Notiz, der gerenderte Bewertungsbericht sowie die Aktionen *Anzeige öffnen*,
*PDF öffnen*, *PDF erzeugen* und *Bericht im Editor*.

**Auswertung** — Kennzahlen, Bewerbungstrichter, Bewertungsverteilung,
Wochenaktivität, Rollentypen, Standorte, Arbeitsmodell und Gehaltsstatistik.

**Werkzeuge** — die Repo-Skripte nach Aufgabe gruppiert, jedes mit deutscher
Beschreibung, Argumentfeld, angezeigter Befehlszeile und mitlaufender Ausgabe.
Werkzeuge, die Dateien verändern, sind als solche gekennzeichnet; nach einem
erfolgreichen Schreiblauf wird die Bewerbungsliste automatisch neu eingelesen.

## Aufbau

| Pfad | Inhalt |
|---|---|
| `main.go` | Einstiegspunkt, `--path`, automatische Verzeichnissuche |
| `internal/app/` | Fenster und Bereiche (`app`, `pipeline`, `detail`, `analytics`, `tools`) |
| `internal/runner/` | Werkzeugkatalog und Prozessausführung |
| `internal/ui/` | Farbschema und Formatierungshelfer |

### Die Datenschicht wird geteilt, nicht kopiert

Das Modul heißt `github.com/santifer/career-ops/dashboard/guiapp` und liegt per
`replace` auf `../dashboard`. Der Modulpfad ist bewusst so gewählt: Go erlaubt
den Import von `dashboard/internal/data` nur aus Paketen, deren Import-Pfad mit
`github.com/santifer/career-ops/dashboard/` beginnt. So nutzen TUI und GUI
denselben Tracker-Parser, dieselbe Metrikberechnung und dieselbe PDF-Auflösung
— eine Änderung dort wirkt in beiden Oberflächen.

Das *Verzeichnis* liegt trotzdem auf Repo-Ebene und nicht unter `dashboard/`,
weil `dashboard/*` laut `DATA_CONTRACT.md` System-Layer ist und von
`update-system.mjs` überschrieben wird.

### Statusänderungen

Geschrieben wird über `data.UpdateApplicationStatus` bzw.
`…AndNotes` aus der geteilten Datenschicht. Die benutzt dieselbe Sperrdatei wie
`set-status.mjs`, sodass eine Änderung in der Oberfläche nicht mit einem
gleichzeitig laufenden Skript kollidiert. Die Statusauswahl bietet nur die
kanonischen Werte aus `templates/states.yml` an; angezeigt werden deutsche
Beschriftungen, geschrieben wird immer der kanonische Wert.

## Tests

```bash
cd gui && go test ./...
```

Abgedeckt sind der Werkzeugkatalog (jeder Eintrag zeigt auf ein real
vorhandenes Skript), die Baumstruktur des Werkzeugbereichs, die
Prozessausführung gegen echte Node-Prozesse (Ausgabe-Streaming, Exit-Codes,
Abbruch, kein Parallelstart) sowie Statusübersetzung und Formatierung.
