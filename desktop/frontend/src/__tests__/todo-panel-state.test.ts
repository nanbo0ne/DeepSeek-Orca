import { createTodoPanelState, isTodoPanelOpen, reduceTodoPanelState } from "../lib/todoPanelState";

let passed = 0;
let failed = 0;

function check(value: boolean, label: string) {
  if (value) {
    process.stdout.write(`  PASS  ${label}\n`);
    passed += 1;
  } else {
    process.stdout.write(`  FAIL  ${label}\n`);
    failed += 1;
  }
}

console.log("\ntodo panel state");
let state = createTodoPanelState("list-a");
check(!isTodoPanelOpen(state), "new lists start collapsed");

state = reduceTodoPanelState(state, { type: "hover", value: true });
check(isTodoPanelOpen(state), "hover opens the panel");
state = reduceTodoPanelState(state, { type: "hover", value: false });
check(!isTodoPanelOpen(state), "leaving closes an unpinned panel");

state = reduceTodoPanelState(state, { type: "focus", value: true });
check(isTodoPanelOpen(state), "keyboard focus opens the panel");
state = reduceTodoPanelState(state, { type: "close" });
check(!isTodoPanelOpen(state), "Escape or outside close collapses the panel");

state = reduceTodoPanelState(state, { type: "toggle-pin" });
state = reduceTodoPanelState(state, { type: "hover", value: false });
check(isTodoPanelOpen(state) && state.pinned, "click pins the panel open");
const sameList = reduceTodoPanelState(state, { type: "list", listId: "list-a" });
check(sameList === state && sameList.pinned, "updates to the same list preserve pinning");
const newList = reduceTodoPanelState(state, { type: "list", listId: "list-b" });
check(!isTodoPanelOpen(newList) && !newList.pinned, "a new list resets to collapsed");

console.log(`\n${passed} passed, ${failed} failed, ${passed + failed} total`);
if (failed > 0) process.exit(1);
