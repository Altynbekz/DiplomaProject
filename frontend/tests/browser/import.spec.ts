import { test, expect } from "@playwright/test";
import { resolve } from "node:path";
import { readFile } from "node:fs/promises";
test("synthetic report persists, deduplicates and renders safely after reload", async ({
  page,
}) => {
  const token = process.env.E2E_API_TOKEN;
  if (!token) throw new Error("E2E_API_TOKEN required");
  const unexpected: string[] = [];
  page.on("request", (r) => {
    const u = new URL(r.url());
    if (!["127.0.0.1", "localhost"].includes(u.hostname))
      unexpected.push(r.url());
  });
  await page.goto("/");
  await page.getByLabel("API token").fill(token);
  await page.getByRole("button", { name: "Unlock", exact: true }).click();
  await expect(page.getByText(/No reports imported/)).toBeVisible();
  await page.getByLabel("Target name").fill("Synthetic browser lab");
  await page.getByRole("button", { name: "Create Juice Shop target" }).click();
  await page.getByRole("link", { name: "Synthetic browser lab" }).click();
  await page
    .getByLabel("Report file")
    .setInputFiles(
      resolve("../backend/testdata/zap/synthetic-traditional.json"),
    );
  await page
    .getByRole("button", { name: "Import report", exact: true })
    .click();
  await expect(
    page.getByRole("status").filter({ hasText: "observations." }),
  ).toContainText("Report persisted.");
  await page.getByText("Explicit private evidence access").click();
  const download = page.getByRole("button", {
    name: "Download original report",
  });
  await expect(download).toBeDisabled();
  await page
    .getByRole("checkbox", {
      name: "I acknowledge this contains private lab data",
    })
    .check();
  const downloadEvent = page.waitForEvent("download");
  await download.click();
  const original = await downloadEvent;
  expect(original.suggestedFilename()).toBe("zap-report.json");
  expect(await readFile(await original.path())).toEqual(
    await readFile(
      resolve("../backend/testdata/zap/synthetic-traditional.json"),
    ),
  );
  const rows = page.locator("tbody tr");
  await expect(rows).toHaveCount(4);
  await expect(rows.first()).toContainText("unreviewed");
  await expect(rows.first()).toContainText("Not assessed");
  await expect(page.getByText("5 / 4 / 1", { exact: true })).toBeVisible();
  await page
    .getByRole("button", { name: "Import report", exact: true })
    .click();
  await expect(
    page.getByRole("status").filter({ hasText: "observations." }),
  ).toContainText("Existing import reused.");
  await expect(rows).toHaveCount(4);
  await page.reload();
  await expect(page.getByLabel("API token")).toHaveValue("");
  await page.getByLabel("API token").fill(token);
  await page.getByRole("button", { name: "Unlock", exact: true }).click();
  await expect(rows).toHaveCount(4);
  await page
    .getByLabel("Scanner severity", { exact: true })
    .selectOption("unknown");
  await expect(rows).toHaveCount(1);
  await expect(rows.first()).toContainText("unknown");
  await page
    .getByLabel("Verification", { exact: true })
    .selectOption("confirmed");
  await expect(
    page.getByText("No findings match these filters."),
  ).toBeVisible();
  await page.getByRole("button", { name: "Clear filters" }).click();
  await expect(rows).toHaveCount(4);
  await page
    .getByRole("link", {
      name: "Synthetic <script>window.pwned=1</script>",
      exact: true,
    })
    .first()
    .click();
  await expect(
    page.getByText("Not assessed / Not assessed", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText(/\/site\/0\/alerts\/0\/instances\//).first(),
  ).toBeVisible();
  await expect(page.locator(".scanner-text").first()).toContainText(
    "<script>window.pwned=1</script>",
  );
  expect(await page.evaluate(() => "pwned" in window)).toBe(false);
  expect(await page.locator("body").innerText()).not.toContain(
    "synthetic-private-value",
  );
  expect(unexpected).toEqual([]);
});

test("persisted rows survive API and database restart", async ({ page }) => {
  const token = process.env.E2E_API_TOKEN;
  if (!token) throw new Error("E2E_API_TOKEN required");
  await page.goto("/");
  await page.getByLabel("API token").fill(token);
  await page.getByRole("button", { name: "Unlock", exact: true }).click();
  await page.getByRole("link", { name: "Synthetic browser lab" }).click();
  await expect(page.locator("tbody tr")).toHaveCount(4);
});
