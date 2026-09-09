// Package runner startet die Node-Werkzeuge des career-ops-Repos und liefert
// deren Ausgabe zeilenweise an die Oberfläche.
package runner

// Tool beschreibt ein aufrufbares Repo-Kommando.
type Tool struct {
	// Titel in der Werkzeugliste.
	Name string
	// Skriptdatei im career-ops-Wurzelverzeichnis, z. B. "scan.mjs".
	Script string
	// Feste Argumente, die immer mitgegeben werden.
	Args []string
	// Deutsche Kurzbeschreibung für die Detailspalte.
	Desc string
	// Beispiel für das Zusatzargument-Feld; leer heißt "keine Argumente nötig".
	ArgHint string
	// Writes markiert Werkzeuge, die Dateien im Repo verändern. Die Oberfläche
	// weist darauf hin, bevor sie laufen.
	Writes bool
}

// Group bündelt Werkzeuge zu einem Themenbereich.
type Group struct {
	Name  string
	Tools []Tool
}

// Catalog ist der kuratierte Werkzeugkatalog: die praktisch nutzbaren
// Kommandos des Repos, gruppiert nach Aufgabe. Bewusst nicht alle ~130
// .mjs-Dateien — viele davon sind Testdateien oder interne Bibliotheken ohne
// sinnvollen Einzelaufruf.
func Catalog() []Group {
	return []Group{
		{
			Name: "Suchen & Scannen",
			Tools: []Tool{
				{
					Name: "Portale scannen", Script: "scan.mjs",
					Desc:    "Durchsucht die in portals.yml eingetragenen Quellen über deren Schnittstellen — welche das sind, steht in Ihrer eigenen Konfiguration (aktuell Arbeitsagentur, Arbeitnow, Deutsche Bahn sowie Suchabfragen für StepStone und Indeed). Insgesamt kennt der Scanner 78 Anbieter. Verursacht keine Modellkosten.",
					ArgHint: "--limit 50",
					Writes:  true,
				},
				{
					Name: "Vollständiger ATS-Scan", Script: "scan-ats-full.mjs",
					Desc:    "Kehrt die Richtung um: statt Ihrer Firmenliste durchsucht er öffentliche Firmenverzeichnisse von Greenhouse, Lever, Ashby, Workday und iCIMS — rund 38.000 Firmen — und behält, was Ihre Titel- und Ortsfilter passieren lässt. Diese fünf Systeme sind stark US-lastig; für den deutschen Markt ist die Ausbeute gering. Läuft Stunden, --resume setzt einen Abbruch fort.",
					ArgHint: "--ats workday,icims --since 7",
					Writes:  true,
				},
				{
					Name: "Hacker-News-Scan", Script: "scan-hn.mjs",
					Desc:   "Wertet den monatlichen \"Who is hiring\"-Thread von Hacker News aus.",
					Writes: true,
				},
				{
					Name: "Interamt scannen", Script: "scan-interamt.mjs",
					Desc:    "Browser-Scanner für Interamt.de, das Stellenportal des deutschen öffentlichen Dienstes.",
					ArgHint: "--query Informatiker",
					Writes:  true,
				},
				{
					Name: "Anzeigen auf Gültigkeit prüfen", Script: "check-liveness.mjs",
					Desc:    "Prüft, ob die verlinkten Stellenanzeigen noch online sind.",
					ArgHint: "optional eine URL",
				},
				{
					Name: "Wiederholte Ausschreibungen finden", Script: "detect-reposts.mjs",
					Desc:    "Markiert Stellen, die binnen 90 Tagen mehrfach neu ausgeschrieben wurden — ein Warnsignal.",
					ArgHint: "--summary",
				},
			},
		},
		{
			Name: "Bewerbungsliste pflegen",
			Tools: []Tool{
				{
					Name: "Übersicht anzeigen", Script: "tracker.mjs",
					Desc: "Zeigt die Bewerbungsliste aus data/applications.md zusammengefasst an.",
				},
				{
					Name: "Eintrag suchen", Script: "find.mjs",
					Desc:    "Sucht in Bewerbungen und Berichten nach einem Stichwort.",
					ArgHint: "Suchbegriff, z. B. Acme",
				},
				{
					Name: "Status setzen", Script: "set-status.mjs",
					Desc:    "Der offizielle Schreibweg für Statusänderungen inklusive Prüfungen und Sperre. Argumente: <Berichtsnummer|Firma> <Status> [--note \"…\"].",
					ArgHint: "42 Applied --note \"per Portal\"",
					Writes:  true,
				},
				{
					Name: "Neue Einträge einpflegen", Script: "merge-tracker.mjs",
					Desc:   "Führt die TSV-Dateien aus batch/tracker-additions/ in die Bewerbungsliste ein. Nach jedem Stapellauf ausführen.",
					Writes: true,
				},
				{
					Name: "Doppelte entfernen", Script: "dedup-tracker.mjs",
					Desc:   "Findet und entfernt doppelte Zeilen in der Bewerbungsliste.",
					Writes: true,
				},
				{
					Name: "Status vereinheitlichen", Script: "normalize-statuses.mjs",
					Desc:   "Setzt abweichende Statuswerte auf die kanonischen Werte aus templates/states.yml zurück.",
					Writes: true,
				},
				{
					Name: "Datenbestand prüfen", Script: "verify-pipeline.mjs",
					Desc: "Gesundheitsprüfung: fehlende Berichte, ungültige Status, kaputte Verweise.",
				},
				{
					Name: "Pipeline abgleichen", Script: "reconcile-pipeline.mjs",
					Desc:   "Gleicht data/pipeline.md gegen die Bewerbungsliste ab.",
					Writes: true,
				},
			},
		},
		{
			Name: "Auswertungen",
			Tools: []Tool{
				{
					Name: "Gesamtstatistik", Script: "stats.mjs",
					Desc:    "Lebenslaufzeit-Zahlen: Trichter, Scan-Summen, Portalabdeckung, Nachfass-Disziplin.",
					ArgHint: "--summary",
				},
				{
					Name: "Muster analysieren", Script: "analyze-patterns.mjs",
					Desc: "Absagemuster und Erfolgsquote je ATS-Anbieter.",
				},
				{
					Name: "Trichtergeschwindigkeit", Script: "funnel-velocity.mjs",
					Desc:    "Vergleicht die eigenen Übergangsquoten mit Marktwerten und misst die Dauer je Stufe.",
					ArgHint: "--summary",
				},
				{
					Name: "Qualifikationslücken", Script: "upskill.mjs",
					Desc: "Gewichtete Übersicht der Fähigkeiten, die in den Ausschreibungen gefordert, aber im Lebenslauf nicht belegt sind.",
				},
				{
					Name: "Gehaltslücke", Script: "salary-gap.mjs",
					Desc:    "Vergleicht Wunschgehalt, ausgeschriebenes und tatsächlich beobachtetes Gehalt.",
					ArgHint: "--summary",
				},
				{
					Name: "Lücken zu einer Stellenanzeige", Script: "jd-skill-gap.mjs",
					Desc:    "Vergleicht eine Stellenbeschreibung gegen cv.md: vorhanden / belegt / Lücke.",
					ArgHint: "Pfad zur JD-Datei",
				},
				{
					Name: "Späte Absagen", Script: "rejection-latency.mjs",
					Desc:    "Findet Firmen, bei denen das Verfahren ungewöhnlich lange stillsteht.",
					ArgHint: "--summary",
				},
				{
					Name: "Prozessqualität", Script: "process-quality.mjs",
					Desc:    "Reibungsrate je Firma aus den Notizen der aktiven Gespräche.",
					ArgHint: "--summary",
				},
				{
					Name: "Firmenhistorie", Script: "company-history.mjs",
					Desc:    "Alles zu einer Firma an einer Stelle: Bewerbungen, Nachfassungen, Scans, Statusverlauf.",
					ArgHint: "Firmenname",
				},
				{
					Name: "Wochenrückblick", Script: "weekly-digest.mjs",
					Desc:    "Fasst die Gesprächsprotokolle der Woche zusammen.",
					ArgHint: "--summary",
				},
			},
		},
		{
			Name: "Unterlagen",
			Tools: []Tool{
				{
					Name: "PDF erzeugen", Script: "generate-pdf.mjs",
					Desc:    "Rendert eine HTML-Fassung des Lebenslaufs über Playwright nach PDF.",
					ArgHint: "<eingabe.html> <ausgabe.pdf>",
					Writes:  true,
				},
				{
					Name: "LaTeX-Lebenslauf bauen", Script: "generate-latex.mjs",
					Desc:    "Prüft die LaTeX-Fassung und übersetzt sie mit pdflatex.",
					ArgHint: "Pfad zur .tex-Datei",
					Writes:  true,
				},
				{
					Name: "Anschreiben erzeugen", Script: "generate-cover-letter.mjs",
					Desc:    "Erstellt ein Anschreiben aus einer Vorgabedatei.",
					ArgHint: "--payload <datei.json>",
					Writes:  true,
				},
				{
					Name: "Lebenslauf als HTML bauen", Script: "build-cv-html.mjs",
					Desc:   "Setzt cv.md mit templates/cv-template.html zu einer HTML-Fassung zusammen.",
					Writes: true,
				},
				{
					Name: "Lebenslauf-Fakten prüfen", Script: "verify-cv-facts.mjs",
					Desc: "Prüft, ob alle Zahlen im erzeugten Lebenslauf auf cv.md zurückgehen. Schutz vor erfundenen Angaben.",
				},
				{
					Name: "Lebenslauf-Abgleich", Script: "cv-sync-check.mjs",
					Desc: "Meldet Abweichungen zwischen cv.md und den daraus erzeugten Fassungen.",
				},
				{
					Name: "Bilder zu PDF", Script: "img-to-pdf.mjs",
					Desc:    "Fasst Bilddateien, etwa Zeugnisscans, zu einer PDF zusammen.",
					ArgHint: "<bild1> <bild2> … <ausgabe.pdf>",
					Writes:  true,
				},
				{
					Name: "Anzeige archivieren", Script: "archive-posting.mjs",
					Desc:    "Sichert eine Stellenanzeige nach jds/, damit sie nach dem Offline-Gehen noch vorliegt. Mit --report=N bleibt sie über die Berichtsnummer auffindbar.",
					ArgHint: "<url> --report=42",
					Writes:  true,
				},
			},
		},
		{
			Name: "Nachfassen & Kontakte",
			Tools: []Tool{
				{
					Name: "Nachfass-Takt berechnen", Script: "followup-cadence.mjs",
					Desc: "Berechnet, bei welchen Bewerbungen eine Nachfrage ansteht.",
				},
				{
					Name: "Erste Nachfassung anlegen", Script: "followup-seed.mjs",
					Desc:   "Trägt für frisch versendete Bewerbungen einen ersten Nachfasstermin ein.",
					Writes: true,
				},
				{
					Name: "Kontakte exportieren", Script: "contacts.mjs",
					Desc:    "Exportiert die Kontaktliste als vCard mit stabilen Kennungen, damit ein erneuter Import nichts doppelt anlegt.",
					ArgHint: "--vcf",
				},
				{
					Name: "Einladung zuordnen", Script: "invite-match.mjs",
					Desc:    "Ordnet eine eingefügte Gesprächseinladung der passenden Bewerbung zu.",
					ArgHint: "--summary",
				},
				{
					Name: "Antwort einfügen", Script: "paste-reply.mjs",
					Desc:    "Nimmt eine per Hand eingefügte E-Mail-Antwort in die Auswertung auf, ohne die Bewerbungsliste zu verändern.",
					ArgHint: "Pfad zur Textdatei",
					Writes:  true,
				},
			},
		},
		{
			Name: "System",
			Tools: []Tool{
				{
					Name: "Einrichtung prüfen", Script: "doctor.mjs",
					Desc:    "Prüft, ob cv.md, config/profile.yml, modes/_profile.md und portals.yml vorhanden sind. Der erste Aufruf bei Problemen.",
					ArgHint: "--json",
				},
				{
					Name: "Auf Aktualisierung prüfen", Script: "update-system.mjs",
					Desc:    "Vergleicht die installierte Fassung mit der Vorlage. Nur \"check\" liest; \"apply --confirm\" schreibt.",
					Args:    []string{"check"},
					ArgHint: "check",
				},
				{
					Name: "Portale prüfen (Konfiguration)", Script: "validate-portals.mjs",
					Desc: "Prüft portals.yml auf Formfehler.",
				},
				{
					Name: "Portale prüfen (erreichbar)", Script: "verify-portals.mjs",
					Desc: "Testet, ob die konfigurierten Portale antworten.",
				},
				{
					Name: "Erweiterungen", Script: "plugins.mjs",
					Desc:    "Listet und aktiviert optionale Erweiterungen.",
					ArgHint: "list",
					Args:    []string{"list"},
				},
				{
					Name: "Merkzettel", Script: "agent-inbox.mjs",
					Desc:    "Offene Notizen für die nächste Sitzung. Nichts wird automatisch ausgeführt.",
					ArgHint: "list",
					Args:    []string{"list"},
				},
				{
					Name: "Ergebnis festhalten", Script: "outcome.mjs",
					Desc:    "Hält den Ausgang einer Bewerbung fest und archiviert die zugehörigen Dateien.",
					ArgHint: "<auswahl> <typ>",
					Writes:  true,
				},
				{
					Name: "Berichtsnummern reservieren", Script: "reserve-report-num.mjs",
					Desc:    "Reserviert einen Nummernblock, bevor mehrere Bewertungen parallel laufen.",
					ArgHint: "--count 4",
					Writes:  true,
				},
				{
					Name: "Eintrag hinzufügen", Script: "add-entry.mjs",
					Desc:    "Nimmt ein abgeschlossenes Projekt oder eine Station in den Lebenslauf auf — mit Vorschau vor dem Schreiben.",
					ArgHint: "--help",
					Writes:  true,
				},
				{
					Name: "Test-Ergebnis protokollieren", Script: "assessment-log.mjs",
					Desc:    "Trägt das Ergebnis eines Fähigkeitstests in data/assessments.tsv ein.",
					ArgHint: "add --platform … --score …",
					Writes:  true,
				},
				{
					Name: "Datenaktualität prüfen", Script: "check-table-freshness.mjs",
					Desc:    "Meldet veraltete Länder- und Rechtstabellen unter templates/.",
					ArgHint: "--summary",
				},
			},
		},
	}
}
