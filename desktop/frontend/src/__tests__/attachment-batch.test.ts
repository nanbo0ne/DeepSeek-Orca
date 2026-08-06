import { DedupIndex } from "../lib/attachDedup";

let passed = 0;
let failed = 0;
function check(value: boolean, label: string) {
  if (value) { console.log(`  PASS  ${label}`); passed += 1; }
  else { console.log(`  FAIL  ${label}`); failed += 1; }
}

console.log("\nattachment batch identity");
const index = new DedupIndex();
index.add("same-hash", "file:one.docx:application/octet-stream:10:1");
check(index.seen("same-hash", "file:one.docx:application/octet-stream:10:1"), "repeated source is deduplicated");
check(!index.seen("same-hash", "file:two.docx:application/octet-stream:10:1"), "same bytes with a different name remain distinct");
index.add("same-hash", "file:two.docx:application/octet-stream:10:1");
check(index.seen("other-hash", "file:two.docx:application/octet-stream:10:1"), "source remains addressable after processing");

console.log(`\n${passed} passed, ${failed} failed, ${passed + failed} total`);
if (failed > 0) process.exit(1);
