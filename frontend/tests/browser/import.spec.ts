import { test, expect } from "@playwright/test";
import { resolve } from "node:path";
import { mkdir, readFile } from "node:fs/promises";
async function capture(page: import("@playwright/test").Page, name: string) {
  const dir = process.env.E2E_SCREENSHOTS;
  if (!dir) return;
  await mkdir(dir, { recursive: true });
  await page.screenshot({ path: resolve(dir, name), fullPage: true });
}

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
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.goto("/");
  await capture(page, "desktop-unlock.png");
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("link", { name: "Skip to content" }),
  ).toBeFocused();
  const tokenField = page.getByLabel("API token");
  await tokenField.fill("invalid-test-token");
  await tokenField.press("Enter");
  await expect(page.getByRole("alert")).toHaveText(
    "Token not accepted. Check it and try again.",
  );
  await expect(
    page.getByRole("heading", { name: "Unlock workspace" }),
  ).toBeVisible();
  await tokenField.fill(token);
  await tokenField.press("Enter");
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
    page.getByRole("status").filter({ hasText: "Report imported." }),
  ).toContainText("Report imported.");
  await expect(
    page.getByText(
      /5 observations · 4 findings · 1 unlocated · Provenance incomplete/,
    ),
  ).toBeVisible();
  await expect(page.locator("tbody tr")).toHaveCount(4);
  await capture(page, "desktop-findings.png");
  await page.getByText("Source report details", { exact: true }).click();
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
  await expect(page.getByText("SHA-256 of uploaded bytes")).toBeVisible();
  await page.getByText("Optional report provenance").click();
  const provenance = page.getByLabel("Provenance JSON");
  await provenance.fill('{"authorization_ref":"synthetic-change"}');
  await page
    .getByRole("button", { name: "Import report", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(
    "stored provenance is immutable",
  );
  await expect(provenance).toHaveValue(
    '{"authorization_ref":"synthetic-change"}',
  );
  await provenance.fill("");
  await page
    .getByRole("button", { name: "Import report", exact: true })
    .click();
  await expect(
    page
      .getByRole("status")
      .filter({ hasText: "This report was already imported." }),
  ).toContainText("This report was already imported.");
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
  await page.getByText("More filters", { exact: true }).click();
  await page
    .getByLabel("Verification", { exact: true })
    .selectOption("confirmed");
  await expect(
    page.getByText("No findings match these filters."),
  ).toBeVisible();
  await page.getByRole("button", { name: "Clear filters" }).click();
  await expect(rows).toHaveCount(4);
  await page
    .getByLabel("Scanner severity", { exact: true })
    .selectOption("high");
  await expect(rows).toHaveCount(3);
  await page.getByText("More filters", { exact: true }).click();
  const filteredURL = page.url();
  await page.setViewportSize({ width: 390, height: 844 });
  const scrollRegion = page.getByRole("region", { name: "Findings" });
  await expect(scrollRegion).toBeVisible();
  expect(
    await scrollRegion.evaluate((el) => el.scrollWidth > el.clientWidth),
  ).toBe(true);
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await capture(page, "mobile-findings.png");
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
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await capture(page, "mobile-evidence.png");
  expect(await page.evaluate(() => "pwned" in window)).toBe(false);
  expect(await page.locator("body").innerText()).not.toContain(
    "synthetic-private-value",
  );
  await page.getByRole("link", { name: "← Findings" }).click();
  await expect(page).toHaveURL(filteredURL);
  await expect(rows).toHaveCount(3);

  // Cursors are signed and can become invalid, for example after token rotation.
  // There may be no other active filter to clear in this case.
  const targetPath = new URL(filteredURL).pathname;
  await page.goto(targetPath + "?cursor=stale-cursor");
  await page.getByLabel("API token").fill(token);
  await page.getByLabel("API token").press("Enter");
  await expect(page.getByRole("alert")).toContainText(
    "Invalid or mismatched cursor",
  );
  await page.getByRole("button", { name: "First page" }).click();
  await expect(page).toHaveURL(new URL(targetPath, filteredURL).href);
  await expect(rows).toHaveCount(4);
  await page.getByRole("button", { name: "Lock workspace" }).click();
  await expect(page.getByLabel("API token")).toHaveValue("");
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
