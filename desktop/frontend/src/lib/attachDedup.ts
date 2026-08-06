// attachDedup centralizes source-level deduplication. Content hashes remain in
// the API for compatibility, but different filenames with equal bytes remain
// distinct inputs.

const HEX = "0123456789abcdef";

function bytesToHex(bytes: Uint8Array): string {
  let out = "";
  for (let i = 0; i < bytes.length; i++) {
    const b = bytes[i];
    out += HEX[(b >> 4) & 0xf] + HEX[b & 0xf];
  }
  return out;
}

// sha256 returns the hex SHA-256 of `blob`. The Web Crypto Subtle API
// is available in Wails' WebView (Chromium / WebKitGTK 4.1+); we
// don't fall back to a JS implementation because a no-op (returning
// "") would silently disable dedup, which is worse than no dedup
// at all. The caller checks the empty-string return and skips the
// dedup step in that case.
export async function sha256(blob: Blob): Promise<string> {
  if (typeof crypto === "undefined" || !crypto.subtle) return "";
  try {
    const buf = await blob.arrayBuffer();
    const digest = await crypto.subtle.digest("SHA-256", buf);
    return bytesToHex(new Uint8Array(digest));
  } catch {
    return "";
  }
}

// DedupIndex tracks source identities for the current composer session.
export class DedupIndex {
  private paths = new Set<string>();

  seen(_hash: string, path: string): boolean {
    return this.paths.has(path);
  }

  add(_hash: string, path: string): void {
    this.paths.add(path);
  }

  forget(_hash: string, path: string): void {
    this.paths.delete(path);
  }

  clear(): void {
    this.paths.clear();
  }
}
