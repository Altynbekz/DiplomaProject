import { expect, type Page } from "@playwright/test";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import type { components } from "../../src/api/schema";

type Snapshot = components["schemas"]["AssessmentSnapshot"];
type AssessmentInput = components["schemas"]["AssessmentInput"];
type Finding = components["schemas"]["Finding"];
type Observation = components["schemas"]["Observation"];
type Verification = components["schemas"]["Verification"];
type Classification = components["schemas"]["ClassificationRevision"];
type Context = components["schemas"]["ContextRevision"];

export const syntheticTitle = "Synthetic <script>window.pwned=1</script>";

export async function openSyntheticFinding(page: Page) {
  await page
    .getByRole("link", { name: syntheticTitle, exact: true })
    .first()
    .click();
}

async function exerciseSyntheticAssessments(page: Page, token: string) {
  const headers = { Authorization: "Bearer " + token };
  const findingPath = "/api/v1" + new URL(page.url()).pathname;
  async function read<T>(path: string): Promise<T> {
    const response = await page.request.get(path, { headers });
    expect(response.status()).toBe(200);
    return (await response.json()) as T;
  }
  async function add<T>(path: string, data: object): Promise<T> {
    const response = await page.request.post(path, { headers, data });
    expect(response.status()).toBe(201);
    return (await response.json()) as T;
  }

  const fixture = JSON.parse(
    await readFile(resolve("../backend/testdata/cvss/first-v4.json"), "utf8"),
  ) as {
    cases: {
      name: string;
      base: Record<string, string>;
      environmental: Record<string, string>;
      score: number;
      vector: string;
    }[];
  };
  const sample = fixture.cases.find(
    (item) => item.name === "paired environmental override",
  );
  const baseSample = fixture.cases.find((item) => item.name === "maximum base");
  if (!sample || !baseSample)
    throw new Error("Missing published reference case");
  expect(sample.base).toEqual(baseSample.base);

  const finding = await read<Finding>(findingPath);
  const observation = (
    await read<{ items: Observation[] }>(findingPath + "/observations")
  ).items[0];
  const caption = "Synthetic browser workflow only";
  const refs = [
    {
      observation_id: observation.id,
      source_pointer: observation.source_pointer,
      note: caption,
    },
  ];
  const review = await add<Verification>(findingPath + "/verifications", {
    state: "confirmed",
    procedure: caption,
    rationale: caption,
    evidence_refs: refs,
    expected_previous_id: finding.verification_review_id,
  });
  const classification = await add<Classification>(
    findingPath + "/classifications",
    {
      status: "mapped",
      category: "A02:2025",
      candidate_categories: ["A02:2025"],
      rationale: caption,
      evidence_refs: refs,
      expected_previous_id: finding.classification.revision_id,
    },
  );
  function rationales(metrics: Record<string, string>) {
    return Object.fromEntries(
      Object.keys(metrics).map((name) => [
        name,
        { reason: caption, evidence_refs: refs, context_factor_refs: [] },
      ]),
    );
  }
  const bInput: AssessmentInput = {
    kind: "B",
    finalize: true,
    base_assessment_id: null,
    expected_verification_review_id: review.id,
    expected_classification_revision_id: classification.id,
    expected_context_revision_ids: [],
    metrics: baseSample.base,
    metric_rationales: rationales(baseSample.base),
    notes: caption,
  };
  const path = findingPath + "/assessments";
  const base = await add<Snapshot>(path, bInput);
  expect(base.score).toBe(baseSample.score);

  const contextPath = "/api/v1/targets/" + finding.target_id + "/contexts";
  const factors = {
    exposure: "restricted",
    required_privileges: "user",
    data_sensitivity: "internal",
    business_criticality: "low",
  };
  const provenance = Object.fromEntries(
    Object.keys(factors).map((factor) => [
      factor,
      {
        rationale: caption,
        sources: [
          {
            document_ref: "synthetic-demo-policy",
            sha256: "1".repeat(64),
            note: caption,
          },
        ],
      },
    ]),
  );
  const contextInput = {
    method: null,
    path: null,
    factors,
    factor_provenance: provenance,
    expected_previous_id: null,
  };
  const original = await add<Context>(contextPath, contextInput);
  const environment = await add<Snapshot>(path, {
    ...bInput,
    kind: "BE",
    base_assessment_id: base.id,
    expected_context_revision_ids: [original.id],
    metrics: sample.environmental,
    metric_rationales: rationales(sample.environmental),
    factor_dispositions: Object.fromEntries(
      Object.keys(factors).map((factor) => [
        factor,
        { reason: caption, metrics: [] },
      ]),
    ),
  });
  expect(environment.score).toBe(sample.score);
  const draft = await add<Snapshot>(path, {
    ...bInput,
    finalize: false,
    metrics: {},
    metric_rationales: {},
  });
  expect(draft.score).toBeNull();
  return { sample, base, original };
}

export async function checkAssessmentHistory(page: Page, token: string) {
  await openSyntheticFinding(page);
  const summary = page.getByText("CVSS assessment history", { exact: true });
  await summary.press("Enter");
  await expect(page.getByText("No assessments saved.")).toBeVisible();
  const { sample, base, original } = await exerciseSyntheticAssessments(
    page,
    token,
  );
  await page.getByRole("link", { name: "← Findings" }).click();
  const assessed = page.getByRole("row").filter({
    has: page.getByRole("link", { name: syntheticTitle, exact: true }),
    hasText: sample.vector,
  });
  await expect(assessed).not.toContainText("(historical)");
  const context = page.getByText("Application context", { exact: true });
  await context.press("Enter");
  const editor = context.locator("xpath=..");
  await editor
    .getByRole("article")
    .filter({ hasText: "Target default · Revision 1" })
    .getByRole("button", { name: "Use this revision" })
    .click();
  await editor
    .getByRole("group", { name: "Exposure", exact: true })
    .getByLabel("Value")
    .selectOption("internal");
  await editor.getByRole("button", { name: "Save context revision" }).click();
  await expect(
    editor.getByText("Context revision saved.", { exact: true }),
  ).toBeVisible();
  await expect(assessed).toContainText(
    "CVSS-B Incomplete / CVSS-BE " + sample.score + " (historical)",
  );
  await openSyntheticFinding(page);
  await page.reload();
  await page.getByLabel("API token").fill(token);
  await page.getByRole("button", { name: "Unlock", exact: true }).click();
  await summary.press("Enter");
  const records = page.getByRole("region", { name: "Stored CVSS assessments" });
  await expect(records.getByRole("article")).toHaveCount(3);
  const draft = records.getByRole("article").first();
  await expect(
    draft.getByRole("heading", { name: "CVSS-B", exact: true }),
  ).toBeVisible();
  await expect(draft.getByText("Incomplete · score unavailable")).toBeVisible();
  const completeBase = records.getByRole("article").filter({
    has: page.getByRole("heading", { name: "CVSS-B", exact: true }),
    hasText: `Score ${base.score?.toFixed(1)}`,
  });
  await expect(completeBase).not.toContainText("Historical (stale)");
  await page.getByText("Human verification", { exact: true }).press("Enter");
  await page.getByLabel("Review state").selectOption("inconclusive");
  await page
    .getByLabel("Verification procedure")
    .fill("Synthetic history check");
  await page
    .getByLabel("Verification rationale")
    .fill("Synthetic history check");
  await page
    .getByRole("group", { name: "Source evidence" })
    .getByRole("checkbox")
    .first()
    .check();
  await page.getByRole("button", { name: "Save verification" }).click();
  await expect(
    page.getByText("Verification saved.", { exact: true }),
  ).toBeVisible();
  await expect(
    completeBase.getByText(
      `Score ${base.score?.toFixed(1)} · Historical (stale)`,
    ),
  ).toBeVisible();
  const environment = records.getByRole("article").filter({
    has: page.getByRole("heading", { name: "CVSS-BE", exact: true }),
  });
  await expect(
    environment.getByText(
      `Score ${sample.score.toFixed(1)} · Historical (stale)`,
    ),
  ).toBeVisible();
  await expect(
    environment.getByText(sample.vector, { exact: true }),
  ).toBeVisible();
  await environment
    .getByText("Assessment evidence and provenance", { exact: true })
    .press("Enter");
  await expect(environment.getByText(base.id, { exact: true })).toBeVisible();
  await expect(
    environment.getByText(original.id, { exact: true }),
  ).toBeVisible();
  await expect(
    environment.getByText("restricted", { exact: true }),
  ).toBeVisible();
  await expect(
    environment.getByText("first-cvss-v4/v1", { exact: false }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
}
