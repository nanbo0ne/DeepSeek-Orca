import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const css = readFileSync(join(root, "styles.css"), "utf8");
const chrome = readFileSync(join(root, "components", "AppChrome.tsx"), "utf8");
const app = readFileSync(join(root, "App.tsx"), "utf8");
const composer = readFileSync(join(root, "components", "Composer.tsx"), "utf8");
const settings = readFileSync(join(root, "components", "SettingsPanel.tsx"), "utf8");
const processCard = readFileSync(join(root, "components", "ProcessCard.tsx"), "utf8");
const transcript = readFileSync(join(root, "components", "Transcript.tsx"), "utf8");
const todoPanel = readFileSync(join(root, "components", "TodoPanel.tsx"), "utf8");
const promptShelf = readFileSync(join(root, "components", "PromptShelf.tsx"), "utf8");
const toolCard = readFileSync(join(root, "components", "ToolCard.tsx"), "utf8");
const controller = readFileSync(join(root, "lib", "useController.ts"), "utf8");

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
check(
  app.includes("app.GetProductCapabilities()") &&
    composer.includes("promptModes.map((mode)") &&
    !composer.includes("PROMPT_MODE_OPTIONS") &&
    !composer.includes("CircleHelp") &&
    !composer.includes("promptModeHelpKey"),
  "prompt modes come from product capabilities without descriptions or a help button",
);
check(
  !settings.includes('<SettingsField label={t("settings.botModel")}') &&
    !settings.includes('<SettingsField label={t("settings.botPromptMode")}') &&
    !settings.includes('<SettingsField label={t("settings.botWorkspaceRoot")}') &&
    settings.includes("productCapabilities.automationWorkspaceEnabled === true"),
  "bot and memory settings follow the product capability boundary",
);
check(
  app.includes('automationConversation ? ["assistant"]') &&
    app.includes("promptModeLocked={automationConversation}") &&
    composer.includes("disabled={disabled || promptModeLocked}"),
  "automation conversations expose a visible but locked assistant mode",
);
const chooser = readFileSync(join(root, "components", "NewSessionChooser.tsx"), "utf8");
const projectTree = readFileSync(join(root, "components", "ProjectTree.tsx"), "utf8").replace(/\r\n/g, "\n");
check(
  chooser.includes('choose("automation", "", "automation")') &&
    projectTree.includes('node.kind === "automation_folder"') &&
    projectTree.includes('node.kind === "automation_topic"'),
  "automation workspace can be created and opened from the desktop tree",
);
check(
  projectTree.includes('scope === "automation"\n        ? []') &&
    projectTree.includes('scope === "automation" ? "'),
  "automation root stays fixed instead of inheriting project rename and color actions",
);
const composerContract = css.slice(css.indexOf("/* Composer responsive contract."));
check(
  [720, 580, 460, 380, 320].every((width, index, widths) => {
    const position = composerContract.indexOf(`@container (max-width: ${width}px)`);
    const previous = index === 0 ? -1 : composerContract.indexOf(`@container (max-width: ${widths[index - 1]}px)`);
    return position > previous;
  }),
  "final Composer breakpoints use one descending 720/580/460/380/320 sequence",
);
check(
  composerContract.includes(".composer-meta__control--model {\n    display: inline-flex;") &&
    composerContract.includes(".composer-enhanced__button svg:last-child {\n    display: none;"),
  "narrow Composer retains model access and collapses mode trigger to one icon",
);
check(
    composer.includes('composer-runstatus__primary--${hasDraftContent ? "send" : "stop"}') &&
    composer.includes("onClick={hasDraftContent ? () => void submit() : handleCancel}") &&
    composer.includes("hasDraftContent && (disabled || submitting || pendingPaste > 0 || !hasSendableContent)") &&
    composer.includes('hasDraftContent ? <ArrowUp size={13} /> : <Square size={10} fill="currentColor" />') &&
    css.includes(".composer-runstatus__primary {\n  --wails-draggable: no-drag;") &&
    css.includes("width: 58px;\n  min-width: 58px;\n  max-width: 58px;\n  height: 26px;") &&
    css.includes(".composer-runstatus__primary--send {") &&
    composerContract.includes(".composer-runstatus__primary {\n    width: 28px;\n    min-width: 28px;\n    max-width: 28px;"),
  "running send replaces stop without changing the primary action geometry",
);
check(
  composer.includes("Plain text always follows the textarea's native paste path") &&
    composer.includes('if (pasted !== "") return;') &&
    !composer.includes("shouldFoldPaste") &&
    !composer.includes("composer__pasted"),
  "plain text paste remains editable text and wins over rich clipboard image hints",
);
check(
  app.includes('target.classList.contains("composer__input")') &&
    app.includes("pasteRequest={composerPasteRequest}") &&
    composer.includes("pasteFromContextMenu") &&
    composer.includes("await attachNativeClipboardImage(true)"),
  "custom context-menu paste delegates images and files to Composer",
);
check(
  css.includes(":root[data-theme-style] .composer__btn--send:disabled") &&
    css.includes("background: var(--control-disabled-bg)") &&
    css.includes("color: var(--control-disabled-fg)"),
  "themed disabled send button stays muted",
);
check(
  composer.includes("!hasSendableContent && !(goalModeOn && !activeGoal)") &&
    !composer.includes("!text.trim() && attachments.length === 0 && workspaceRefs.length === 0"),
  "failed attachments cannot leave the ordinary send button falsely enabled",
);
check(
  composer.includes("const COMPOSER_AUTO_MAX_LINES = 10") &&
    composer.includes("composerAutoInputMaxHeight(node)") &&
    composer.includes("node.scrollHeight > maxHeight + 1") &&
    css.includes(".composer__input {\n  flex: 1;\n  resize: none;\n  margin: 0;\n  padding: 0;"),
  "Composer grows with content up to a stable line limit before scrolling",
);
check(
  settings.includes('["compact", "detailed"]') &&
    !settings.includes('["compact", "standard", "detailed"]'),
  "process settings expose only compact and detailed modes",
);
check(
  processCard.includes("const closeFromKeyboard") &&
    processCard.includes("{hasBody && actualOpen && (") &&
    transcript.includes("processOpenOverrides.get(segment.id)") &&
    transcript.includes("next.set(segment.id, nextOpen)"),
  "compact process details can close cleanly and preserve stable segment overrides",
);
check(
  css.includes(".process-activity-mark") &&
    css.includes("prefers-reduced-motion: reduce") &&
    css.includes("process-activity-mark--tool") &&
    css.includes("animation-direction: reverse") &&
    transcript.includes("activityIndicatorPhase(items, activityIndicatorEnabled, running, paused)") &&
    transcript.includes("timeline-entry--activity"),
  "optional activity spinner is direction-aware, reduced-motion safe, and follows the current turn",
);
check(
  css.includes(".footer {\n  position: relative;") &&
    css.includes("border-top: 0;\n  background: transparent;") &&
    css.includes(".footer-shelves") &&
    css.includes("background: transparent;"),
  "footer shelves remain transparent outside their individual cards",
);
check(
  !todoPanel.includes("AnchoredPopover") &&
    todoPanel.includes('className="todobar__surface"') &&
    todoPanel.includes('data-ui-surface="panel"') &&
    todoPanel.includes('className="todobar__details"'),
  "Todo expands inside one in-flow footer surface",
);
check(
  promptShelf.includes('data-ui-surface="panel"') &&
    transcript.includes('data-ui-surface="panel"') &&
    !transcript.includes('className={`turn-stats-row turn-process-panel${open ? " turn-stats-row--open" : ""}`} data-ui-surface="panel"') &&
    !processCard.includes('data-ui-surface="panel"') &&
    !toolCard.includes('data-ui-surface="panel"'),
  "panel ownership stays on the shelf or turn while process and tool rows remain flat",
);
check(
  css.includes(".turn-process-panel {\n  width: min(100%, 820px);\n  overflow: visible;\n  border: 0;\n  border-radius: 0;\n  background: transparent;") &&
    !css.includes(".turn-process-panel > button"),
  "completed-turn summary uses the original transparent inline treatment",
);
check(
  toolCard.includes('data-ui-surface="content"') &&
    css.includes(":root[data-theme-style] .process-activity-rail .process-card__body {\n  padding: 0 4px 8px 25px;\n  border-top: 0;\n  background: transparent;"),
  "only tool output surfaces retain contained backgrounds inside the activity rail",
);
check(
  transcript.includes("transcript--hydrating") &&
    css.includes(".transcript--hydrating .timeline-entry") &&
    css.includes("animation: none !important;"),
  "restored history skips bulk entrance animation during its first paint",
);
check(
  controller.includes("Meta and history are the only first-paint dependencies") &&
    controller.includes("afterNextPaint()") &&
    controller.includes('dispatchTo(meta.id, { type: "session_load_start"') &&
    controller.indexOf("safe(app.EffortForTab(tabId))") > controller.indexOf('type: "session_primary_loaded"'),
  "conversation selection paints before history work and auxiliary status hydrates later",
);

console.log(`\n${passed} passed, ${failed} failed, ${passed + failed} total`);
if (failed > 0) process.exit(1);
