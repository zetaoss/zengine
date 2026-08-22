import { readFileSync, writeFileSync } from "node:fs";

import {
  classifyExtensions,
  DOCKERFILE,
  fetchBaseExtensions,
  readLocalExtensions,
  readZbaseVersion,
  renderGeneratedBlock,
  replaceGeneratedBlock,
} from "./extensions-lib.mjs";

const local = readLocalExtensions();
const zbaseVersion = readZbaseVersion();
const base = await fetchBaseExtensions(zbaseVersion);
const { overrides, downgrades } = classifyExtensions(local, base);

if (downgrades.length > 0) {
  for (const { entry, baseEntry } of downgrades) {
    console.error(`${entry.name}: ${entry.tag} is below zbase ${baseEntry.tag}`);
  }
  process.exit(1);
}

const source = readFileSync(DOCKERFILE, "utf8");
const updated = replaceGeneratedBlock(source, renderGeneratedBlock(overrides));
if (updated !== source) writeFileSync(DOCKERFILE, updated);

const names = overrides.map(({ name }) => name).join(", ") || "none";
console.log(`✅  Docker extension overrides synced (zbase ${zbaseVersion}: ${names})`);
