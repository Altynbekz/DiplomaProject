import { expect, type Locator, type Page } from "@playwright/test";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import type { components } from "../../src/api/schema";

type Snapshot = components["schemas"]["AssessmentSnapshot"];
type Finding = components["schemas"]["Finding"];
type Observation = components["schemas"]["Observation"];
type Verification = components["schemas"]["Verification"];
type Classification = components["schemas"]["ClassificationRevision"];
type Context = components["schemas"]["ContextRevision"];

async function saveFromEditor(page: Page, button: Locator, status: number) {
  const pending = page.waitForResponse(
    (response) =>
      new URL(response.url()).pathname.endsWith("/assessments") &&
      response.request().method() === "POST",
  );
  await button.press("Enter");
  const response = await pending;
  expect(response.status(), await response.text()).toBe(status);
  return response;
}

function metric(editor: Locator, name: string) {
  return editor.getByRole("group", { name: new RegExp(`\\(${name}\\)$`) });
}

async function selectMetric(
  editor: Locator,
  name: string,
  value: string,
  reason: string,
  source: string,
) {
  const field = metric(editor, name);
  await field.getByLabel("Value").selectOption(value);
  await field.getByLabel("Rationale").fill(reason);
  const evidence = field.getByText(/^Observation evidence/);
  if (!((await field.locator("details").getAttribute("open")) !== null)) {
    await evidence.press("Enter");
  }
  await field.getByRole("checkbox", { name: new RegExp(source) }).check();
}

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
  const newBase = page.getByRole("button", { name: "New CVSS-B assessment" });
  await newBase.press("Enter");
  const editor = page.getByRole("region", { name: "CVSS assessment editor" });
  await expect(
    editor.getByRole("heading", { name: "CVSS-B editor" }),
  ).toBeFocused();
  await expect(metric(editor, "AV").getByLabel("Value")).toHaveValue("");
  await expect(editor.getByText(/^Missing Base values/)).toBeVisible();
  await metric(editor, "AV")
    .getByLabel("Value")
    .selectOption(baseSample.base.AV);
  const finishBase = editor.getByRole("button", {
    name: "Finalize CVSS-B",
    exact: true,
  });
  await saveFromEditor(page, finishBase, 422);
  await expect(editor.getByRole("alert")).toContainText(
    "Finalization requires a metric rationale",
  );
  await selectMetric(
    editor,
    "AV",
    baseSample.base.AV,
    caption,
    observation.source_pointer,
  );
  await editor.getByLabel("Assessment notes").fill(caption);
  const firstSave = await saveFromEditor(
    page,
    editor.getByRole("button", { name: "Save incomplete assessment" }),
    201,
  );
  const firstDraft = (await firstSave.json()) as Snapshot;
  expect(firstDraft.metrics).toEqual({ AV: baseSample.base.AV });
  expect(firstDraft.metric_rationales.AV.evidence_refs[0].observation_id).toBe(
    observation.id,
  );
  expect(firstDraft.score).toBeNull();
  expect(firstDraft.vector).toBeNull();

  await add<Verification>(findingPath + "/verifications", {
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
  await page
    .getByRole("button", { name: "Revise as new snapshot" })
    .press("Enter");
  await expect(metric(editor, "AV").getByLabel("Rationale")).toHaveValue(
    caption,
  );
  for (const [name, value] of Object.entries(baseSample.base)) {
    await selectMetric(
      editor,
      name,
      value,
      caption,
      observation.source_pointer,
    );
  }
  await saveFromEditor(page, finishBase, 409);
  await expect(editor.getByRole("alert")).toContainText(
    "The current verification review changed",
  );
  await editor
    .getByRole("button", { name: "Load current assessment inputs" })
    .press("Enter");
  await expect(
    page.getByRole("status").filter({ hasText: "Your draft is retained" }),
  ).toBeVisible();
  await editor
    .getByText("Review and classification revisions", { exact: true })
    .press("Enter");
  await expect(
    editor.getByText("Current verification", { exact: true }),
  ).toBeVisible();
  // An in-page revision changes while assessment recovery is waiting for adoption.
  const reviewPanel = page.getByText("Human verification", { exact: true });
  await reviewPanel.press("Enter");
  const reviewEditor = reviewPanel.locator("xpath=..");
  await reviewEditor.getByLabel("Review state").selectOption("confirmed");
  await reviewEditor.getByLabel("Verification procedure").fill(caption);
  await reviewEditor.getByLabel("Verification rationale").fill(caption);
  await reviewEditor
    .getByRole("group", { name: "Source evidence" })
    .getByRole("checkbox")
    .first()
    .check();
  const reviewSave = reviewEditor.getByRole("button", {
    name: "Save verification",
  });
  const reviewPost = (response: import("@playwright/test").Response) =>
    new URL(response.url()).pathname.endsWith("/verifications") &&
    response.request().method() === "POST";
  const firstAttempt = page.waitForResponse(reviewPost);
  await reviewSave.press("Enter");
  let reviewed = await firstAttempt;
  if (reviewed.status() === 409) {
    await reviewEditor
      .getByRole("button", { name: "Load latest verification" })
      .press("Enter");
    await expect(
      reviewEditor.getByText("Review your draft before saving", {
        exact: false,
      }),
    ).toBeVisible();
    const secondAttempt = page.waitForResponse(reviewPost);
    await reviewSave.press("Enter");
    reviewed = await secondAttempt;
  }
  expect(reviewed.status(), await reviewed.text()).toBe(201);
  const received = (await reviewed.json()) as Verification;
  await expect(editor.getByText(received.id, { exact: true })).toBeVisible();
  await expect(
    editor.getByRole("button", { name: "Use loaded revisions" }),
  ).not.toBeVisible();
  await editor
    .getByRole("button", { name: "Load current assessment inputs" })
    .press("Enter");
  await reviewPanel.press("Enter");
  await editor
    .getByRole("button", { name: "Use loaded revisions" })
    .press("Enter");
  await expect(metric(editor, "SA").getByLabel("Rationale")).toHaveValue(
    caption,
  );
  await expect(
    metric(editor, "AV").getByRole("checkbox", {
      name: new RegExp(observation.source_pointer),
    }),
  ).toBeChecked();
  const base = (await (
    await saveFromEditor(page, finishBase, 201)
  ).json()) as Snapshot;
  expect(base.score).toBe(baseSample.score);
  expect(base.vector).toBe(baseSample.vector);
  expect(base.verification_review_id).toBe(received.id);
  expect(base.id).not.toBe(firstDraft.id);
  expect(
    await read<Snapshot>("/api/v1/assessments/" + firstDraft.id),
  ).toMatchObject({
    status: "incomplete",
    score: null,
    vector: null,
    metrics: firstDraft.metrics,
  });

  const records = page.getByRole("region", { name: "Stored CVSS assessments" });
  const baseRecord = records.getByRole("article").filter({
    has: page.getByText(baseSample.vector, { exact: true }),
  });
  await baseRecord
    .getByRole("button", { name: "Start CVSS-BE from this Base" })
    .press("Enter");
  await expect(
    editor.getByRole("heading", { name: "CVSS-BE editor" }),
  ).toBeFocused();
  await expect(editor.getByText(base.id, { exact: true })).toBeVisible();
  await editor
    .getByText("Inherited Base values", { exact: true })
    .press("Enter");
  await expect(
    editor.getByRole("definition").filter({ hasText: "Network (N)" }),
  ).toBeVisible();
  await expect(metric(editor, "MAV").locator("option").first()).toHaveText(
    "Not Defined (X) · inherits AV:N",
  );
  await expect(metric(editor, "CR").locator("option").first()).toHaveText(
    "Not Defined (X) · CVSS default High",
  );
  const finishBE = editor.getByRole("button", {
    name: "Finalize CVSS-BE",
    exact: true,
  });
  await expect(finishBE).toBeDisabled();
  await editor
    .getByRole("button", { name: "Use displayed context" })
    .press("Enter");
  await expect(editor.getByText(/^Unknown context factors/)).toBeVisible();

  const override = metric(editor, "MAC");
  await override.getByLabel("Value").selectOption(sample.environmental.MAC);
  await override.getByLabel("Rationale").fill(caption);
  await override
    .getByRole("checkbox", { name: "Exposure", exact: true })
    .check();
  await saveFromEditor(page, finishBE, 422);
  await expect(editor.getByRole("alert")).toContainText(
    "Document this factor's disposition",
  );
  for (const factor of [
    "Exposure",
    "Required privileges",
    "Data sensitivity",
    "Business criticality",
  ]) {
    await editor
      .getByLabel(new RegExp("^" + factor + " disposition"))
      .fill(
        factor === "Exposure" ? caption : "Synthetic only: no change justified",
      );
  }
  await saveFromEditor(page, finishBE, 422);
  await expect(editor.getByRole("alert")).toContainText(
    "Review the unknown context factor",
  );
  const draftSave = await saveFromEditor(
    page,
    editor.getByRole("button", { name: "Save incomplete assessment" }),
    201,
  );
  const beDraft = (await draftSave.json()) as Snapshot;
  expect(beDraft).toMatchObject({
    status: "incomplete",
    score: null,
    vector: null,
    base_assessment_id: base.id,
    context_snapshot: { source_revision_ids: [] },
    metrics: {
      ...base.metrics,
      MAC: sample.environmental.MAC,
      MAV: "X",
      CR: "X",
    },
    factor_dispositions: { exposure: { metrics: ["MAC"], reason: caption } },
  });

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
  const beRecord = records.getByRole("article").filter({
    has: page.getByRole("heading", { name: "CVSS-BE", exact: true }),
  });
  await beRecord
    .getByRole("button", { name: "Revise as new snapshot" })
    .press("Enter");
  await expect(
    override.getByRole("checkbox", { name: "Exposure", exact: true }),
  ).toBeChecked();
  for (const [name, value] of Object.entries(sample.environmental)) {
    if (name !== "MAC")
      await selectMetric(
        editor,
        name,
        value,
        caption,
        observation.source_pointer,
      );
  }
  await editor
    .getByLabel("Assessment notes")
    .fill("Synthetic BE draft retained through conflicts");

  async function recover(expected: string, updateRevisions = true) {
    await saveFromEditor(page, finishBE, 409);
    await expect(editor.getByRole("alert")).toContainText(expected);
    await editor
      .getByRole("button", { name: "Load current assessment inputs" })
      .press("Enter");
    await expect(
      page.getByRole("status").filter({ hasText: "Your draft is retained" }),
    ).toBeVisible();
    if (updateRevisions) {
      await editor
        .getByRole("button", { name: "Use loaded revisions" })
        .press("Enter");
    }
    await expect(editor.getByLabel("Assessment notes")).toHaveValue(
      "Synthetic BE draft retained through conflicts",
    );
    await expect(
      override.getByRole("checkbox", { name: "Exposure", exact: true }),
    ).toBeChecked();
  }
  await recover("The resolved context changed");
  await expect(editor.getByText(original.id, { exact: true })).toBeVisible();
  const nextClassification = await add<Classification>(
    findingPath + "/classifications",
    {
      status: "mapped",
      category: "A02:2025",
      candidate_categories: ["A02:2025"],
      rationale: "Synthetic classification revision",
      evidence_refs: refs,
      expected_previous_id: classification.id,
    },
  );
  await recover("The current classification revision changed");
  await recover(
    "The complete Base assessment or its review and classification changed",
    false,
  );

  await baseRecord
    .getByRole("button", { name: "Revise as new snapshot" })
    .press("Enter");
  await expect(
    editor.getByRole("heading", { name: "CVSS-B editor" }),
  ).toBeFocused();
  await editor
    .getByRole("button", { name: "Load current assessment inputs" })
    .press("Enter");
  await expect(
    page.getByRole("status").filter({ hasText: "Your draft is retained" }),
  ).toBeVisible();
  await editor
    .getByRole("button", { name: "Use loaded revisions" })
    .press("Enter");
  const replacementBase = (await (
    await saveFromEditor(page, finishBase, 201)
  ).json()) as Snapshot;
  expect(replacementBase.id).not.toBe(base.id);
  expect(replacementBase.classification_revision_id).toBe(
    nextClassification.id,
  );
  expect(replacementBase.metrics).toEqual(base.metrics);
  expect(replacementBase.metric_rationales).toEqual(base.metric_rationales);

  await page
    .getByRole("button", { name: "Resume CVSS-BE assessment" })
    .press("Enter");
  await expect(editor.getByLabel("Assessment notes")).toHaveValue(
    "Synthetic BE draft retained through conflicts",
  );
  await records
    .getByRole("button", { name: "Use this Base as parent" })
    .press("Enter");
  await expect(
    editor.getByText(replacementBase.id, { exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  const environment = (await (
    await saveFromEditor(page, finishBE, 201)
  ).json()) as Snapshot;
  expect(environment).toMatchObject({
    status: "complete",
    score: sample.score,
    vector: sample.vector,
    base_assessment_id: replacementBase.id,
    classification_revision_id: nextClassification.id,
    context_snapshot: { source_revision_ids: [original.id] },
    notes: "Synthetic BE draft retained through conflicts",
    metrics: { ...replacementBase.metrics, ...sample.environmental, CR: "X" },
  });
  for (const [name, rationale] of Object.entries(
    replacementBase.metric_rationales,
  )) {
    expect(environment.metric_rationales[name]).toEqual(rationale);
  }
  expect(
    await read<Snapshot>("/api/v1/assessments/" + beDraft.id),
  ).toMatchObject({
    status: "incomplete",
    score: null,
    vector: null,
    base_assessment_id: base.id,
    context_snapshot: beDraft.context_snapshot,
  });
  await newBase.press("Enter");
  const latestDraftSave = await saveFromEditor(
    page,
    editor.getByRole("button", { name: "Save incomplete assessment" }),
    201,
  );
  const draft = (await latestDraftSave.json()) as Snapshot;
  expect(draft.score).toBeNull();
  return { sample, base: replacementBase, original };
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
  await expect(records.getByRole("article")).toHaveCount(6);
  const draft = records.getByRole("article").first();
  await expect(
    draft.getByRole("heading", { name: "CVSS-B", exact: true }),
  ).toBeVisible();
  await expect(draft.getByText("Incomplete · score unavailable")).toBeVisible();
  const completeBase = records
    .getByRole("article")
    .filter({
      has: page.getByRole("heading", { name: "CVSS-B", exact: true }),
      hasText: `Score ${base.score?.toFixed(1)}`,
    })
    .first();
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
    has: page.getByText(sample.vector, { exact: true }),
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
  await environment
    .getByText("Paired Base assessment", { exact: true })
    .press("Enter");
  const paired = environment.getByRole("region", {
    name: "Paired Base",
    exact: true,
  });
  await expect(
    paired.getByText("CVSS-B score", { exact: false }),
  ).toContainText(base.score!.toFixed(1));
  await expect(paired.getByText(base.vector!, { exact: true })).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
}
