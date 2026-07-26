import { readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = fileURLToPath(new URL("..", import.meta.url));
const VERSION_FILE = resolve(ROOT, "svelte/src/shared/version.json");
const TARGETS = [
  resolve(ROOT, "mwz/skins/ZetaSkin/skin.json"),
  resolve(ROOT, "mwz/extensions/ZetaExtension/extension.json"),
];

const { gitVersion } = JSON.parse(readFileSync(VERSION_FILE, "utf-8"));

if (typeof gitVersion !== "string" || gitVersion.length === 0) {
  throw new Error(`${VERSION_FILE}: gitVersion must be a non-empty string`);
}

const version = gitVersion.replace(/^v/, "");

for (const file of TARGETS) {
  const source = readFileSync(file, "utf-8");
  const updated = source.replace(
    /^(\s*"version"\s*:\s*)"[^"]*"/m,
    `$1"${version}"`,
  );

  if (updated === source) {
    const currentVersion = JSON.parse(source).version;
    if (currentVersion === version) {
      continue;
    }
    throw new Error(`${file}: version field was not found`);
  }

  writeFileSync(file, updated, "utf-8");
}
