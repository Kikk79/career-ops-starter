// Tests for which CV template a web `pdf` run fills.
//
// The web hardcoded templates/cv-template.html and told the agent so, which meant
// a user's cv.template choice in config/profile.yml applied to every CLI run and
// silently to no web run. These assert the resolution, the fallbacks, and the one
// rule that is a security boundary rather than a cosmetic one: a configured name
// reaches a filename, so it must be normalized, never traversed.
//
// Run:  node --test tests/lib/cv-template.test.mjs

import { test } from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { BASE_TEMPLATE, kebab, templatePathFor, readConfiguredTemplate, resolveCvTemplate } from "../../src/lib/cv-template.mjs";
import { buildPrompt } from "../../src/lib/run-prompts.mjs";

const USABLE = "<html>{{NAME}} {{EXPERIENCE}} {{EDUCATION}}</html>";

/** A throwaway checkout: optional profile.yml, optional template files. */
function makeRoot({ profile, templates = {} } = {}) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "co-cv-template-"));
  if (profile !== undefined) {
    fs.mkdirSync(path.join(root, "config"), { recursive: true });
    fs.writeFileSync(path.join(root, "config", "profile.yml"), profile, "utf8");
  }
  fs.mkdirSync(path.join(root, "templates"), { recursive: true });
  for (const [file, body] of Object.entries(templates)) {
    fs.writeFileSync(path.join(root, "templates", file), body, "utf8");
  }
  return root;
}

test("kebab normalizes a name instead of letting it traverse", () => {
  // Given the name comes from the user's own YAML but reaches a filename
  // Then separators are flattened, never preserved
  assert.equal(kebab("Tuerkis Sidebar"), "tuerkis-sidebar");
  assert.equal(kebab("  MODERN  "), "modern");
  assert.equal(kebab("../../etc/passwd"), "etc-passwd");
  assert.equal(kebab("..\\..\\windows\\system32"), "windows-system32");
  assert.equal(kebab(undefined), "");
});

test("templatePathFor: standard and the empty name are the base file", () => {
  assert.equal(templatePathFor("standard"), BASE_TEMPLATE);
  assert.equal(templatePathFor(""), BASE_TEMPLATE);
  assert.equal(templatePathFor("tuerkis-sidebar"), "templates/cv-template.tuerkis-sidebar.html");
});

test("readConfiguredTemplate: reads cv.template, tolerates every absence", () => {
  assert.equal(readConfiguredTemplate(makeRoot({ profile: "cv:\n  template: tuerkis-sidebar\n" })), "tuerkis-sidebar");
  // Unset key, no cv block, no file at all, and unparseable YAML all mean "default"
  assert.equal(readConfiguredTemplate(makeRoot({ profile: "cv:\n  output_format: html\n" })), null);
  assert.equal(readConfiguredTemplate(makeRoot({ profile: "candidate:\n  full_name: x\n" })), null);
  assert.equal(readConfiguredTemplate(makeRoot()), null);
  assert.equal(readConfiguredTemplate(makeRoot({ profile: "cv:\n  template: [unclosed\n" })), null);
  // A blank value is not a choice
  assert.equal(readConfiguredTemplate(makeRoot({ profile: 'cv:\n  template: "   "\n' })), null);
});

test("resolveCvTemplate: a configured, existing, usable template is used", () => {
  const root = makeRoot({
    profile: "cv:\n  template: tuerkis-sidebar\n",
    templates: { "cv-template.tuerkis-sidebar.html": USABLE },
  });

  const choice = resolveCvTemplate(root);

  assert.equal(choice.path, "templates/cv-template.tuerkis-sidebar.html");
  assert.equal(choice.name, "tuerkis-sidebar");
  assert.deepEqual(choice.warnings, []);
});

test("resolveCvTemplate: no configuration is the base template, silently", () => {
  // Given the base template is the documented default, not a fallback
  // Then it produces NO warning — a warning on the happy path trains users to
  // ignore warnings.
  for (const root of [makeRoot(), makeRoot({ profile: "cv:\n  output_format: html\n" })]) {
    const choice = resolveCvTemplate(root);
    assert.equal(choice.path, BASE_TEMPLATE);
    assert.equal(choice.name, "standard");
    assert.deepEqual(choice.warnings, []);
  }
});

test("resolveCvTemplate: a missing template falls back AND says so", () => {
  const root = makeRoot({ profile: "cv:\n  template: nope\n" });

  const choice = resolveCvTemplate(root);

  assert.equal(choice.path, BASE_TEMPLATE);
  assert.equal(choice.warnings.length, 1);
  // The message has to name both the choice and the file, or the user cannot fix it
  assert.match(choice.warnings[0], /nope/);
  assert.match(choice.warnings[0], /templates\/cv-template\.nope\.html/);
});

test("resolveCvTemplate: a file missing required placeholders is refused", () => {
  // Given a file that exists but cannot be filled — handing it to the agent
  // produces a confident-looking CV with no experience section
  const root = makeRoot({
    profile: "cv:\n  template: broken\n",
    templates: { "cv-template.broken.html": "<html>{{NAME}}</html>" },
  });

  const choice = resolveCvTemplate(root);

  assert.equal(choice.path, BASE_TEMPLATE);
  assert.equal(choice.warnings.length, 1);
  assert.match(choice.warnings[0], /\{\{EXPERIENCE\}\}/);
  assert.match(choice.warnings[0], /\{\{EDUCATION\}\}/);
});

test("resolveCvTemplate: a traversing name cannot escape templates/", () => {
  // Given a hand-edited profile.yml, and a real file planted outside templates/
  const root = makeRoot({ profile: 'cv:\n  template: "../../secret"\n' });
  fs.writeFileSync(path.join(root, "secret.html"), USABLE, "utf8");

  const choice = resolveCvTemplate(root);

  // Then the name is flattened, resolves to a nonexistent templates/ file, and
  // falls back — the planted file is never reachable
  assert.equal(choice.path, BASE_TEMPLATE);
  assert.match(choice.warnings[0] ?? "", /templates\/cv-template\.secret\.html/);
});

test("buildPrompt: the pdf prompt names the resolved template, base by default", () => {
  const ARGS = { input: "018", memory: "", today: "2026-08-31" };

  // Given a resolved template, the prompt names THAT file...
  const tailored = buildPrompt({ kind: "pdf", ...ARGS, cvTemplate: "templates/cv-template.tuerkis-sidebar.html" });
  assert.match(tailored, /templates\/cv-template\.tuerkis-sidebar\.html/);
  // ...and no longer the old unconditional claim that web runs use the base one
  assert.ok(!/web runs always use the base template/i.test(tailored));

  // ...and with none resolved it is the base template, exactly as before
  assert.match(buildPrompt({ kind: "pdf", ...ARGS }), /templates\/cv-template\.html/);
});
