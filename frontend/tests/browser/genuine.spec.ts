import { expect, test, type Page } from "@playwright/test";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const name = "Genuine Juice Shop fixture";
const fixture = resolve("../backend/testdata/zap/juice-shop-traditional.json");

async function unlock(page: Page) {
  const token = process.env.E2E_API_TOKEN;
  if (!token) throw new Error("E2E_API_TOKEN required");
  await page.getByLabel("API token").fill(token);
  await page.getByRole("button", { name: "Unlock", exact: true }).click();
}

async function checkFindings(page: Page, state = "unreviewed") {
  const rows = page.locator("tbody tr");
  await expect(rows).toHaveCount(17);
  await expect(rows.first()).toContainText(state);
  await expect(rows.first()).toContainText("Not assessed");
  await rows.first().getByRole("link").click();
  await expect(
    page.getByText(/\/site\/0\/alerts\/\d+\/instances\/\d+/).first(),
  ).toBeVisible();
}

async function checkRootContext(page: Page) {
  await page
    .getByRole("link", { name: "Content Security Policy (CSP) Header Not Set" })
    .click();
  await page.getByText("Resolved context", { exact: true }).click();
  const context = page.getByRole("region", {
    name: "Resolved context factors",
  });
  await expect(context.getByText("restricted", { exact: true })).toBeVisible();
  await expect(
    context.getByText("synthetic-browser-policy", { exact: false }),
  ).toBeVisible();
  await expect(context.getByText("unknown", { exact: true })).toHaveCount(3);
}

async function prepareClassification(
  page: Page,
  status: "unmapped" | "out_of_scope",
  rationale: string,
) {
  const summary = page.getByText("OWASP classification", { exact: true });
  await summary.press("Enter");
  const panel = summary.locator("xpath=..");
  await panel.getByLabel("Classification status").selectOption(status);
  await panel.getByLabel("Classification rationale").fill(rationale);
  const observation = panel
    .getByRole("group", { name: "Source evidence" })
    .getByRole("checkbox")
    .first();
  await observation.press("Space");
  await expect(observation).toBeChecked();
  return panel;
}

async function saveAndCheckContext(page: Page) {
  const summary = page.getByText("Application context", { exact: true });
  await summary.focus();
  await page.keyboard.press("Enter");
  const panel = summary.locator("xpath=..");
  const scope = panel.getByLabel("Context scope");
  const exposure = panel.getByRole("group", { name: "Exposure", exact: true });
  for (const dimension of [
    "Exposure",
    "Required privileges",
    "Data sensitivity",
    "Business criticality",
  ]) {
    await panel
      .getByRole("group", { name: dimension, exact: true })
      .getByLabel("Rationale")
      .fill("Unknown in the disposable browser test");
  }
  const save = panel.getByRole("button", { name: "Save context revision" });
  const status = panel
    .getByRole("status")
    .filter({ hasText: "Context revision saved." });
  let finishRequest!: () => void;
  const delayed = new Promise<void>((resolve) => {
    finishRequest = resolve;
  });
  const mutationPath = "**/api/v1/targets/*/contexts";
  await page.route(mutationPath, async (route) => {
    if (route.request().method() === "POST") await delayed;
    await route.continue();
  });
  await save.click();
  await expect(scope).toBeDisabled();
  await expect(exposure.getByLabel("Rationale")).toBeDisabled();
  finishRequest();
  await expect(status).toBeVisible();
  await page.unroute(mutationPath);
  await expect(
    panel.getByRole("heading", { name: "Target default · Revision 1" }),
  ).toBeVisible();
  await scope.selectOption("endpoint");
  await scope.selectOption("default");
  await exposure
    .getByLabel("Rationale")
    .fill("Retain this updated synthetic rationale");
  await save.click();
  await expect(panel.getByRole("alert")).toContainText(
    "The current revision changed",
  );
  const prior = panel
    .getByRole("article")
    .filter({ hasText: "Target default · Revision 1" });
  await prior.getByRole("button", { name: "Use this revision" }).click();
  await expect(exposure.getByLabel("Rationale")).toHaveValue(
    "Retain this updated synthetic rationale",
  );
  await save.click();
  await expect(
    panel.getByRole("heading", { name: "Target default · Revision 2" }),
  ).toBeVisible();
  await scope.selectOption("endpoint");
  await panel.getByLabel("Endpoint path").fill("/");
  await exposure.getByLabel("Value").selectOption("restricted");
  await exposure
    .getByLabel("Rationale")
    .fill("Synthetic context value for browser verification");
  await exposure.getByRole("button", { name: "Add lab document" }).click();
  await exposure
    .getByLabel("Lab document reference")
    .fill("synthetic-browser-policy");
  await exposure.getByLabel("Document SHA-256").fill("1".repeat(64));
  await save.click();
  await expect(
    panel.getByRole("heading", { name: "GET / · Revision 1" }),
  ).toBeVisible();
  await scope.selectOption("default");
  await scope.selectOption("endpoint");
  await exposure.getByLabel("Rationale").fill("Keep this endpoint draft");
  await save.click();
  await expect(panel.getByRole("alert")).toContainText(
    "The current revision changed",
  );
  await expect(
    panel
      .getByRole("article")
      .filter({ hasText: "Target default · Revision 2" })
      .getByRole("button", { name: "Use this revision" }),
  ).toBeDisabled();
  await panel
    .getByRole("article")
    .filter({ hasText: "GET / · Revision 1" })
    .getByRole("button", { name: "Use this revision" })
    .click();
  await expect(scope).toHaveValue("endpoint");
  await expect(panel.getByLabel("Endpoint path")).toHaveValue("/");
  await expect(exposure.getByLabel("Rationale")).toHaveValue(
    "Keep this endpoint draft",
  );
  await save.click();
  await expect(
    panel.getByRole("heading", { name: "GET / · Revision 2" }),
  ).toBeVisible();
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(panel.getByLabel("Context scope")).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await checkRootContext(page);
  await page.getByRole("link", { name: "← Findings" }).click();
}

test("genuine report survives reload", async ({ page }) => {
  const external: string[] = [];
  page.on("request", (request) => {
    const url = new URL(request.url());
    if (url.hostname !== "127.0.0.1") external.push(url.origin);
  });
  await page.goto("/");
  await unlock(page);
  await page.getByLabel("Target name").fill(name);
  await page.getByRole("button", { name: "Create Juice Shop target" }).click();
  await page.getByRole("link", { name }).click();
  await page.getByLabel("Report file").setInputFiles(fixture);
  await page.getByText("Optional report provenance").click();
  const provenance = await readFile(
    resolve("../backend/testdata/zap/juice-shop-provenance.json"),
    "utf8",
  );
  await page.getByLabel("Provenance JSON").fill(provenance);
  const submit = page.getByRole("button", {
    name: "Import report",
    exact: true,
  });
  await submit.click();
  await expect(
    page.getByRole("status").filter({ hasText: "Report imported." }),
  ).toBeVisible();
  await expect(
    page.getByText("21 observations · 17 findings", { exact: true }),
  ).toBeVisible();
  await submit.click();
  await expect(
    page
      .getByRole("status")
      .filter({ hasText: "This report was already imported." }),
  ).toBeVisible();
  await page.getByText("Source report details", { exact: true }).click();
  await page.getByText("Recorded provenance (null means unknown)").click();
  await expect(
    page.getByText(JSON.parse(provenance).authorization_ref, { exact: false }),
  ).toBeVisible();
  await page.reload();
  await unlock(page);
  await saveAndCheckContext(page);
  await checkFindings(page);
  const stale = await page.context().newPage();
  await stale.goto(page.url());
  await unlock(stale);
  await stale.getByText("Human verification", { exact: true }).click();
  await expect(stale.getByText("No human reviews yet.")).toBeVisible();
  await stale
    .getByRole("combobox", { name: "Review state" })
    .selectOption("inconclusive");
  await stale
    .getByLabel("Verification procedure")
    .fill("Check the scanner assertion independently");
  await stale
    .getByLabel("Verification rationale")
    .fill("Keep the draft when a predecessor changes");
  await stale
    .getByRole("group", { name: "Source evidence" })
    .getByRole("checkbox")
    .first()
    .check();
  await page.getByText("Human verification", { exact: true }).click();
  await expect(page.getByText("No human reviews yet.")).toBeVisible();
  const review = page.getByRole("combobox", { name: "Review state" });
  await review.selectOption("inconclusive");
  await page
    .getByLabel("Verification procedure")
    .fill("Inspect the stored test observation");
  await page
    .getByLabel("Verification rationale")
    .fill("The scanner assertion needs human review");
  await page
    .getByRole("group", { name: "Source evidence" })
    .getByRole("checkbox")
    .first()
    .check();
  await page.getByRole("button", { name: "Save verification" }).click();
  await expect(
    page.getByRole("status").filter({ hasText: "Verification saved." }),
  ).toBeVisible();
  await expect(
    page
      .getByRole("region", { name: "Verification history" })
      .getByText("Inspect the stored test observation"),
  ).toBeVisible();
  await stale.getByRole("button", { name: "Save verification" }).click();
  await expect(stale.getByRole("alert")).toContainText(
    "The current revision changed",
  );
  await stale.getByRole("button", { name: "Load latest verification" }).click();
  await expect(
    stale.getByRole("status").filter({ hasText: "Latest revision loaded." }),
  ).toBeVisible();
  await expect(stale.getByLabel("Verification rationale")).toHaveValue(
    "Keep the draft when a predecessor changes",
  );
  await stale.getByRole("button", { name: "Save verification" }).click();
  await expect(
    stale.getByRole("status").filter({ hasText: "Verification saved." }),
  ).toBeVisible();
  await expect(
    stale
      .getByRole("region", { name: "Verification history" })
      .getByText("Check the scanner assertion independently"),
  ).toBeVisible();
  const classification = await prepareClassification(
    page,
    "out_of_scope",
    "Keep the classification history for the offline fixture",
  );
  const staleClassification = await prepareClassification(
    stale,
    "unmapped",
    "Keep this classification draft after a conflict",
  );
  await classification
    .getByRole("button", { name: "Save classification" })
    .press("Enter");
  await expect(
    classification
      .getByRole("status")
      .filter({ hasText: "Classification saved." }),
  ).toBeVisible();
  await expect(
    classification
      .getByRole("region", { name: "Classification history" })
      .getByText("Keep the classification history for the offline fixture"),
  ).toBeVisible();
  const saveClassification = staleClassification.getByRole("button", {
    name: "Save classification",
  });
  await saveClassification.press("Enter");
  await expect(staleClassification.getByRole("alert")).toContainText(
    "The current revision changed",
  );
  await staleClassification
    .getByRole("button", { name: "Load latest classification" })
    .press("Enter");
  await expect(
    staleClassification
      .getByRole("status")
      .filter({ hasText: "Latest classification loaded." }),
  ).toBeVisible();
  await expect(
    staleClassification.getByLabel("Classification rationale"),
  ).toHaveValue("Keep this classification draft after a conflict");
  await expect(
    staleClassification.getByLabel("Classification status"),
  ).toHaveValue("unmapped");
  await saveClassification.press("Enter");
  await expect(
    staleClassification
      .getByRole("region", { name: "Classification history" })
      .getByText("Keep this classification draft after a conflict"),
  ).toBeVisible();
  await stale.close();
  expect(external).toEqual([]);
});

test("genuine findings survive restart", async ({ page }) => {
  await page.goto("/");
  await unlock(page);
  await page.getByRole("link", { name }).click();
  await checkFindings(page, "inconclusive");
  await page.getByText("OWASP classification", { exact: true }).click();
  await expect(
    page
      .getByRole("region", { name: "Classification history" })
      .getByText("Keep the classification history for the offline fixture"),
  ).toBeVisible();
  await expect(
    page
      .getByRole("region", { name: "Classification history" })
      .getByText("Keep this classification draft after a conflict"),
  ).toBeVisible();
  await page.getByText("Human verification", { exact: true }).click();
  await expect(
    page
      .getByRole("region", { name: "Verification history" })
      .getByText("Inspect the stored test observation"),
  ).toBeVisible();
  await page.getByRole("link", { name: "← Findings" }).click();
  await checkRootContext(page);
});
