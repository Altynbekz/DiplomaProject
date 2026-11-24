// Run with a clean, pinned FIRST checkout. The app and its Go library are
// deliberately absent from this process.
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { createContext, runInContext } from "node:vm";

const referenceCommit = "c5b0d409ae9f57c44264c6ce5f27d89298e1d32a";
const dir = process.argv[2];
if (!dir) throw new Error("Supply the pinned FIRST reference checkout path");
const git = (...args) => execFileSync("git", ["-C", dir, ...args], { encoding: "utf8" }).trim();
if (git("rev-parse", "HEAD") !== referenceCommit || git("status", "--porcelain")) {
  throw new Error("The FIRST reference checkout must be clean and at the pinned commit");
}

const sandbox = createContext({});
for (const file of ["metrics.js", "cvss_lookup.js", "max_composed.js", "max_severity.js", "cvss_score.js"]) {
  runInContext(readFileSync(join(dir, file), "utf8"), sandbox, { filename: file });
}
const { expectedMetricOrder } = sandbox;
const baseOrder = Object.keys(expectedMetricOrder).slice(0, 11);
const max = { AV: "N", AC: "L", AT: "N", PR: "N", UI: "N", VC: "H", VI: "H", VA: "H", SC: "H", SI: "H", SA: "H" };
const zero = { ...max, VC: "N", VI: "N", VA: "N", SC: "N", SI: "N", SA: "N" };
const partialImpact = { ...max, AV: "A", AT: "P", PR: "L", UI: "P", VC: "L", VI: "N", VA: "L", SC: "N", SI: "N", SA: "N" };

const inputs = [
  ["maximum base", max, null],
  ["zero base", zero, null],
  ["mixed base and reversed input order", Object.fromEntries(Object.entries(partialImpact).reverse()), null],
  ["physical access and active interaction", { ...max, AV: "P", AC: "H", UI: "A", SC: "L", SI: "L", SA: "N" }, null],
  ["environmental requirements and inheritance", partialImpact, { AR: "L", IR: "M", CR: "H" }],
  ["explicit Not Defined and override", max, { MVA: "X", MAV: "P", CR: "X" }],
  ["modified subsequent integrity safety", partialImpact, { MSI: "S", IR: "H" }],
  ["modified subsequent availability safety", partialImpact, { MSA: "S", AR: "H" }],
  ["modified impacts reduce effective score to zero", max, { MVC: "N", MVI: "N", MVA: "N", MSC: "N", MSI: "N", MSA: "N" }],
  ["modified impact supplies effective impact", zero, { MVC: "L", CR: "M", MPR: "H" }],
  ["paired environmental override", max, { MAT: "P", MAC: "H", MPR: "L", MUI: "A" }],
  ["reference rounding Base regression", { ...max, AT: "P", PR: "H", UI: "A", VC: "L", VI: "L", VA: "L" }, null],
  ["reference rounding BE regression", { ...max, AV: "A", AT: "P", PR: "L", UI: "A", VC: "L", VI: "L", VA: "L", SI: "N" }, { MAV: "L", MPR: "N", MSI: "H" }],
];

const cases = inputs.map(([name, base, environmental]) => {
  const selected = Object.fromEntries(Object.keys(expectedMetricOrder).map(key => [key, "X"]));
  for (const [key, value] of Object.entries({ ...base, ...environmental })) {
    if (!expectedMetricOrder[key]?.includes(value)) throw new Error("Invalid fixture input");
    selected[key] = value;
  }
  const vector = "CVSS:4.0" + Object.entries(selected)
    .filter(([key, value]) => baseOrder.includes(key) || value !== "X")
    .map(([key, value]) => "/" + key + ":" + value).join("");
  sandbox.selected = selected;
  const score = Number(runInContext("cvss_score(selected, cvssLookup_global, maxSeverity, macroVector(selected))", sandbox));
  if (!Number.isFinite(score)) throw new Error("Reference did not produce a score");
  return { name, base, environmental, vector, score };
});

process.stdout.write(JSON.stringify({
  version: "first-cvss-v4/v1",
  reference: "https://github.com/FIRSTdotorg/cvss-v4-calculator",
  reference_commit: referenceCommit,
  reference_license: "BSD-2-Clause",
  cases,
}, null, 2) + "\n");
