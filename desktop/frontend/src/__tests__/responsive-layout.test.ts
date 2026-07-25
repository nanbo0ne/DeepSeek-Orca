import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const css = readFileSync(join(root, "styles.css"), "utf8");
const chrome = readFileSync(join(root, "components", "AppChrome.tsx"), "utf8");
const app = readFileSync(join(root, "App.tsx"), "utf8");
const settings = readFileSync(join(root, "components", "SettingsPanel.tsx"), "utf8");

let passed = 0;
let failed = 0;
function check(value: boolean, label: string) {
  if (value) { process.stdout.write(`  PASS  ${label}\n`); passed += 1; }
  else { process.stdout.write(`  FAIL  ${label}\n`); failed += 1; }
}

console.log("\nresponsive layout safeguards");
check(!css.includes("padding: 5px 120px"), "composer has no fixed right-side reservation");
check(!css.includes(".composer-card__actions {\n    right:"), "composer actions stay in grid flow");
check(chrome.includes("app-chrome__action--core") && chrome.includes("app-chrome__action--tertiary"), "chrome actions define shrink priorities");
check(css.includes("@media (max-width: 1100px)") && css.includes("@media (max-width: 780px)"), "chrome priorities have narrow breakpoints");
check(app.includes("topicbar__overflow-menu") && css.includes(".topicbar__action--direct-utility"), "topic actions expose a narrow overflow menu");
check(css.includes(".statusbar {\n    max-width: none;\n    gap: 6px;\n    overflow: hidden"), "narrow status bar cannot wrap or overflow");
check(
  css.includes(".footer-shelves > .todobar {\n  flex: 0 0 auto;\n  width: min(520px, 100%);") &&
    css.includes("min-width: min(300px, 100%);\n  max-width: 100%;\n  margin-inline: auto;"),
  "Todo stays centered and constrained by the conversation footer",
);
check(
  settings.includes('<SettingsField label={t("settings.autoCheckUpdates")}') &&
    settings.includes('<div className="settings-update-control">') &&
    !settings.includes('<SettingsField label={t("settings.checkUpdatesNow")}'),
  "manual update check stays beside the automatic update toggle",
);

console.log(`\n${passed} passed, ${failed} failed, ${passed + failed} total`);
if (failed > 0) process.exit(1);
