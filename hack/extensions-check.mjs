import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

import {
  classifyExtensions,
  compareVersions,
  DOCKERFILE,
  fetchBaseExtensions,
  readGeneratedBlock,
  readLocalExtensions,
  readZbaseVersion,
  renderGeneratedBlock,
  ROOT,
} from "./extensions-lib.mjs";

function installedVersion(name) {
  const extensionJson = resolve(ROOT, "w/extensions", name, "extension.json");
  let info;
  try {
    info = JSON.parse(readFileSync(extensionJson, "utf8"));
  } catch (error) {
    const reason = error.code === "ENOENT" ? "not installed" : `invalid extension.json (${error.message})`;
    return { error: `${name}: ${reason}` };
  }
  if (info.name !== name) return { error: `${name}: extension.json name is ${JSON.stringify(info.name)}` };
  if (typeof info.version !== "string") return { error: `${name}: extension.json has no version` };
  return { version: info.version };
}

function generatedExtensionNames(block) {
  const match = block.match(/^    && rm -rf (.+) \\$/m);
  return match ? [...match[1].matchAll(/'([^']+)'/g)].map(([, name]) => name) : [];
}

const local = readLocalExtensions();
const zbaseVersion = readZbaseVersion();
const base = await fetchBaseExtensions(zbaseVersion);
const { overrides, downgrades } = classifyExtensions(local, base);
const errors = [];
const extensionsDir = resolve(ROOT, "w/extensions");

if (existsSync(extensionsDir)) {
  for (const entry of local) {
    const installed = installedVersion(entry.name);
    if (installed.error) {
      errors.push(installed.error);
    } else if (compareVersions(installed.version, entry.tag) !== 0) {
      errors.push(`${entry.name}: installed ${installed.version}, expected ${entry.tag}`);
    }
  }
}

for (const { entry, baseEntry } of downgrades) {
  errors.push(`${entry.name}: ${entry.tag} is below zbase ${baseEntry.tag}`);
}

const expectedBlock = renderGeneratedBlock(overrides);
const actualBlock = readGeneratedBlock(readFileSync(DOCKERFILE, "utf8"));
if (actualBlock !== expectedBlock) {
  if (overrides.length === 0) {
    const names = generatedExtensionNames(actualBlock).join(", ") || "none";
    errors.push(
      `Dockerfile overrides out of date: ${names} not needed; run \`make extensions-sync\``,
    );
  } else {
    const names = overrides.map(({ name, tag }) => `${name}@${tag}`).join(", ");
    errors.push(
      `Dockerfile is missing or has outdated generated overrides for ${names}; run \`make extensions-sync\` to update them`,
    );
  }
}

if (errors.length > 0) {
  console.error("Extension check failed:");
  for (const error of errors) console.error(`- ${error}`);
  process.exitCode = 1;
} else {
  console.log(`✅  Extensions checked (${local.length} configured; ${overrides.length} override zbase ${zbaseVersion})`);
}
