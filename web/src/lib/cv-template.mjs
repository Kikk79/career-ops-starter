/**
 * cv-template.mjs — which CV template a web `pdf` run fills.
 *
 * The web used to hardcode `templates/cv-template.html` and say so in the prompt
 * ("web runs always use the base template"), so a user who had picked a template
 * via `cv.template` in config/profile.yml got it on every CLI run and silently
 * never on a web run. This resolves the same key, against the same naming
 * convention cv-templates.mjs documents.
 *
 * It deliberately does NOT import cv-templates.mjs: that module resolves against
 * its own `__dirname`, while the web resolves against `careerOpsRoot()`, which
 * CAREER_OPS_ROOT can point at a different checkout entirely.
 *
 * Plain .mjs (same pattern as pdf-paths.mjs / cv-envelope.mjs) so
 * tests/lib/cv-template.test.mjs can import it under `node --test`.
 */
import fs from "node:fs";
import path from "node:path";
import * as yaml from "js-yaml";

/** Repo-relative, POSIX-separated — it goes into a prompt, not into path.join. */
export const BASE_TEMPLATE = "templates/cv-template.html";

/**
 * Placeholders cv-templates.mjs requires of a `cv` template. A file missing any
 * of them is not a CV template, whatever it is named — falling back beats handing
 * the agent a document it cannot fill.
 */
const REQUIRED_PLACEHOLDERS = ["NAME", "EXPERIENCE", "EDUCATION"];

/**
 * Normalize a template name to the kebab-case form the filenames use.
 *
 * This is the security boundary, not a cosmetic step: the name comes from the
 * user's own YAML but reaches a filename, so `../../etc/passwd` must become
 * `etc-passwd` rather than traverse. Same transformation as cv-templates.mjs's
 * `kebab`, kept identical so web and CLI resolve one name to one file.
 *
 * @param {unknown} display
 * @returns {string}
 */
export function kebab(display) {
  return String(display ?? "")
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

/**
 * The repo-relative template path for a template name.
 * "standard" (and the empty name) is the base file; anything else is the
 * `cv-template.<name>.html` variant.
 *
 * @param {string} name - Already kebab-normalized.
 * @returns {string}
 */
export function templatePathFor(name) {
  return !name || name === "standard" ? BASE_TEMPLATE : `templates/cv-template.${name}.html`;
}

/**
 * Read `cv.template` out of a checkout's config/profile.yml.
 *
 * @param {string} root - careerOpsRoot().
 * @returns {string|null} The raw configured name, or null when unset/unreadable.
 */
export function readConfiguredTemplate(root) {
  let doc;
  try {
    // js-yaml's load() uses the safe default schema — same call pdf-paths.mjs and
    // api/profile/route.ts already make against this file.
    doc = yaml.load(fs.readFileSync(path.join(root, "config", "profile.yml"), "utf8"));
  } catch {
    // Missing or unparseable profile: the base template is the documented default.
    return null;
  }
  const name = doc?.cv?.template;
  return typeof name === "string" && name.trim() ? name.trim() : null;
}

/**
 * @typedef {Object} CvTemplateChoice
 * @property {string} path - Repo-relative POSIX path to hand the agent.
 * @property {string} name - The resolved template name ("standard" for the base file).
 * @property {string[]} warnings - Non-fatal issues to surface in the run log.
 */

/**
 * Resolve the CV template a pdf run should fill, falling back to the base
 * template when the configured one is missing or unusable.
 *
 * Falls back rather than failing — a broken template name should not block a CV
 * — but every fallback produces a warning, because a silent one means the user
 * gets an unexpected layout with nothing anywhere saying why.
 *
 * @param {string} root - careerOpsRoot().
 * @returns {CvTemplateChoice}
 */
export function resolveCvTemplate(root) {
  const configured = readConfiguredTemplate(root);
  if (!configured) return { path: BASE_TEMPLATE, name: "standard", warnings: [] };

  const name = kebab(configured);
  if (!name || name === "standard") return { path: BASE_TEMPLATE, name: "standard", warnings: [] };

  const rel = templatePathFor(name);
  const abs = path.join(root, ...rel.split("/"));
  const fallback = (why) => ({ path: BASE_TEMPLATE, name: "standard", warnings: [why] });

  let text;
  try {
    text = fs.readFileSync(abs, "utf8");
  } catch {
    return fallback(`config/profile.yml selects the CV template "${configured}", but ${rel} does not exist — using the base template.`);
  }

  const missing = REQUIRED_PLACEHOLDERS.filter((ph) => !text.includes(`{{${ph}}}`));
  if (missing.length) {
    return fallback(
      `${rel} is missing required placeholders (${missing.map((m) => `{{${m}}}`).join(", ")}) — using the base template.`,
    );
  }

  return { path: rel, name, warnings: [] };
}
