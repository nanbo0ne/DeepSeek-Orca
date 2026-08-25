import { createRequire } from "node:module";
import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const requireFromSite = createRequire(resolve(root, "site/package.json"));
const sharp = requireFromSite("sharp");

const sourceMaster = resolve(root, "assets/brand/orca-source.png");
const source = resolve(process.argv[2] || sourceMaster);
const symbolTarget = resolve(root, "desktop/frontend/src/assets/logo-symbol.png");
const wordmarkTarget = resolve(root, "desktop/frontend/src/assets/logo-wordmark.png");
const appIconTarget = resolve(root, "desktop/build/appicon.png");
const icoTarget = resolve(root, "desktop/build/windows/icon.ico");
const legacyLogoTarget = resolve(root, "desktop/frontend/src/assets/logo.png");

const transparent = { r: 0, g: 0, b: 0, alpha: 0 };
// The source already has a real alpha channel. Extract the symbol's visible
// area, then place that untouched artwork on a larger transparent square.
// This keeps the ring and whale balanced without including the wordmark.
const symbolExtract = { left: 123, top: 69, width: 1007, height: 1001 };
const symbolCanvasSize = 1120;

async function cleanSymbol(size) {
  const extracted = await sharp(source)
    .extract(symbolExtract)
    .png()
    .toBuffer();
  const padded = await sharp(extracted)
    .extend({
      top: Math.floor((symbolCanvasSize - symbolExtract.height) / 2),
      bottom: Math.ceil((symbolCanvasSize - symbolExtract.height) / 2),
      left: Math.floor((symbolCanvasSize - symbolExtract.width) / 2),
      right: Math.ceil((symbolCanvasSize - symbolExtract.width) / 2),
      background: transparent,
    })
    .png()
    .toBuffer();
  return sharp(padded)
    .resize(size, size, { fit: "fill", kernel: sharp.kernel.lanczos3 })
    .png({ compressionLevel: 9, adaptiveFiltering: true })
    .toBuffer();
}

async function cleanWordmark() {
  const extracted = await sharp(source)
    .extract({ left: 170, top: 1088, width: 914, height: 112 })
    .png()
    .toBuffer();
  return sharp(extracted)
    .trim({ background: transparent, threshold: 18 })
    .resize({ width: 164, height: 20, fit: "inside", withoutEnlargement: true, kernel: sharp.kernel.lanczos3 })
    .png({ compressionLevel: 9, adaptiveFiltering: true })
    .toBuffer();
}

function encodeICO(images) {
  const directorySize = 6 + images.length * 16;
  const header = Buffer.alloc(directorySize);
  header.writeUInt16LE(0, 0);
  header.writeUInt16LE(1, 2);
  header.writeUInt16LE(images.length, 4);
  let offset = directorySize;
  images.forEach(({ size, data }, index) => {
    const at = 6 + index * 16;
    header.writeUInt8(size >= 256 ? 0 : size, at);
    header.writeUInt8(size >= 256 ? 0 : size, at + 1);
    header.writeUInt8(0, at + 2);
    header.writeUInt8(0, at + 3);
    header.writeUInt16LE(1, at + 4);
    header.writeUInt16LE(32, at + 6);
    header.writeUInt32LE(data.length, at + 8);
    header.writeUInt32LE(offset, at + 12);
    offset += data.length;
  });
  return Buffer.concat([header, ...images.map(({ data }) => data)]);
}

const sourceBytes = await readFile(source);
await mkdir(dirname(sourceMaster), { recursive: true });
if (source !== sourceMaster) {
  await writeFile(sourceMaster, sourceBytes);
}
await Promise.all([symbolTarget, wordmarkTarget, appIconTarget, icoTarget].map((target) => mkdir(dirname(target), { recursive: true })));

const symbol = await cleanSymbol(1024);
const wordmark = await cleanWordmark();
const sizes = [16, 20, 24, 32, 48, 64, 128, 256];
const icoImages = await Promise.all(sizes.map(async (size) => ({ size, data: await cleanSymbol(size) })));

await Promise.all([
  writeFile(symbolTarget, symbol),
  writeFile(appIconTarget, symbol),
  writeFile(wordmarkTarget, wordmark),
  writeFile(icoTarget, encodeICO(icoImages)),
]);
await rm(legacyLogoTarget, { force: true });

process.stdout.write(`Generated transparent O.R.C.A assets from ${source}\n`);
