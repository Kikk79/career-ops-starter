import { NextRequest } from "next/server";
import fs from "node:fs";
import path from "node:path";
import { careerOpsRoot, findReportFile } from "@/lib/career-ops";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

const escapeRe = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

/**
 * The company slug the WRITER used, derived the same way resolvePdfPaths does:
 * from the report's FILENAME, not from the tracker's company name. The two
 * routinely differ — reports/009-fritzmeier-….md against a tracker row reading
 * "Georg Fritzmeier GmbH" — and matching on the tracker name alone 404s on a PDF
 * that rendered perfectly well.
 */
function reportCompanySlug(n: string): string | null {
  if (!/^\d+$/.test(n)) return null;
  const reportFile = findReportFile(n);
  if (!reportFile) return null;
  return path.basename(reportFile).match(/^\d+-(.+)-\d{4}-\d{2}-\d{2}\.md$/)?.[1] ?? null;
}

// Serve the tailored CV PDF the pdf mode wrote to output/cv-…-{company}-…pdf for
// a given offer. Inline so it opens in the browser. Local-first: reads the user's
// own output/ dir.
export async function GET(req: NextRequest) {
  const n = (req.nextUrl.searchParams.get("n") ?? "").trim();
  const company = (req.nextUrl.searchParams.get("company") ?? "").trim();
  if (!n && !company) return new Response("n or company required", { status: 400 });

  const dir = path.join(careerOpsRoot(), "output");
  let all: string[];
  try {
    all = fs.readdirSync(dir).filter((f) => f.toLowerCase().endsWith(".pdf"));
  } catch {
    return new Response("no output directory", { status: 404 });
  }

  const slug = reportCompanySlug(n);
  let files: string[] = [];
  if (slug) {
    // The exact shape the backend writes: cv-{candidate}-{slug}-{YYYY-MM-DD}.pdf.
    // Anchored on the date rather than a wildcard, so report 009 ("fritzmeier")
    // cannot serve report 010's ("fritzmeier-netzwerk") CV — and so there is no
    // `.*` to backtrack over.
    const dated = new RegExp(`-${escapeRe(slug)}-\\d{4}-\\d{2}-\\d{2}\\.pdf$`, "i");
    files = all.filter((f) => f.toLowerCase().startsWith("cv-") && dated.test(f));
    if (!files.length) {
      // A CLI-generated CV carries no date segment. Token boundary so "meta"
      // doesn't serve "metabase".
      const loose = new RegExp(`(^|[^a-z0-9])${escapeRe(slug)}([^a-z0-9]|$)`, "i");
      files = all.filter((f) => loose.test(f.toLowerCase()));
    }
  }
  if (!files.length && company) {
    // Fallback for links that carry no report number (older tabs) and for offers
    // whose report cannot be resolved. Token-extract instead of replace-then-trim:
    // same slug, and no `-+$`-style pattern that backtracks polynomially (CodeQL).
    const companySlug = (company.toLowerCase().match(/[a-z0-9]+/g) ?? []).join("-");
    const re = new RegExp(`(^|[^a-z0-9])${escapeRe(companySlug)}([^a-z0-9]|$)`, "i");
    files = all.filter((f) => re.test(f.toLowerCase()));
  }
  if (!files.length) return new Response("no tailored CV found for this offer", { status: 404 });

  files.sort((a, b) => fs.statSync(path.join(dir, b)).mtimeMs - fs.statSync(path.join(dir, a)).mtimeMs);
  const file = path.join(dir, files[0]);
  try {
    const buf = fs.readFileSync(file);
    return new Response(new Uint8Array(buf), {
      status: 200,
      headers: { "Content-Type": "application/pdf", "Content-Disposition": `inline; filename="${files[0]}"`, "Cache-Control": "no-store" },
    });
  } catch {
    return new Response("could not read the PDF", { status: 500 });
  }
}
