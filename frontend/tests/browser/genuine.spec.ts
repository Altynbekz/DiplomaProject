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

async function checkFindings(page: Page) {
  const rows = page.locator("tbody tr");
  await expect(rows).toHaveCount(17);
  await expect(rows.first()).toContainText("unreviewed");
  await expect(rows.first()).toContainText("Not assessed");
  await rows.first().getByRole("link").click();
  await expect(
    page.getByText(/\/site\/0\/alerts\/\d+\/instances\/\d+/).first(),
  ).toBeVisible();
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
  await checkFindings(page);
  expect(external).toEqual([]);
});

test("genuine findings survive restart", async ({ page }) => {
  await page.goto("/");
  await unlock(page);
  await page.getByRole("link", { name }).click();
  await checkFindings(page);
});
