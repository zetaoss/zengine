import { existsSync, rmSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { resolve } from "node:path";

import { readLocalExtensions, ROOT } from "./extensions-lib.mjs";

const EXTENSIONS_DIR = resolve(ROOT, "w/extensions");
const entries = readLocalExtensions();
for (const entry of entries) {
  const target = resolve(EXTENSIONS_DIR, entry.name);
  if (!target.startsWith(`${EXTENSIONS_DIR}/`)) {
    throw new Error(`Refusing to remove extension outside ${EXTENSIONS_DIR}: ${entry.name}`);
  }
  if (existsSync(target)) {
    console.log(`Removing ${target}`);
    rmSync(target, { recursive: true, force: true });
  }

  console.log(`Cloning ${entry.name} (${entry.tag})`);
  const result = spawnSync(
    "git",
    ["clone", "--branch", entry.tag, entry.repo, target],
    { stdio: "inherit" },
  );
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}

console.log(`✅  Re-cloned ${entries.length} extensions`);
