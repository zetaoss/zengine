import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";

export const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..");
export const DOCKERFILE = resolve(ROOT, "Dockerfile");
export const LOCAL_EXTENSIONS = resolve(ROOT, "hack/extensions.yaml");

const ZIMAGE_RAW_URL = "https://raw.githubusercontent.com/zetaoss/zimage";
const GENERATED_START = "# extensions-sync: generated; do not edit";
const GENERATED_END = "# /extensions-sync: generated; do not edit";

export function parseExtensions(yaml, source) {
  const entries = [];
  let entry;

  for (const [index, rawLine] of yaml.split("\n").entries()) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) continue;

    const match = line.match(/^(?:- )?(name|repo|tag):\s*(.*?)\s*$/);
    if (!match) throw new Error(`${source}:${index + 1}: unsupported YAML syntax`);

    const [, key, rawValue] = match;
    const value = rawValue.replace(/^(?:"(.*)"|'(.*)')$/, "$1$2");
    if (key === "name") {
      if (entry) entries.push(entry);
      entry = { name: value };
      continue;
    }
    if (!entry) throw new Error(`${source}:${index + 1}: ${key} precedes name`);
    entry[key] = value;
  }

  if (entry) entries.push(entry);
  const names = new Set();
  for (const extension of entries) {
    for (const key of ["name", "repo", "tag"]) {
      if (!extension[key]) throw new Error(`${source}: missing ${key}`);
    }
    if (!/^[A-Za-z0-9._-]+$/.test(extension.name)) {
      throw new Error(`${source}: invalid extension name ${JSON.stringify(extension.name)}`);
    }
    if (names.has(extension.name)) throw new Error(`${source}: duplicate extension ${extension.name}`);
    names.add(extension.name);
    parseVersion(extension.tag);
  }
  return entries;
}

export function readLocalExtensions() {
  return parseExtensions(readFileSync(LOCAL_EXTENSIONS, "utf8"), LOCAL_EXTENSIONS);
}

export function readZbaseVersion() {
  const dockerfile = readFileSync(DOCKERFILE, "utf8");
  const matches = [...dockerfile.matchAll(/^\s*ARG\s+ZBASE_VERSION\s*=\s*([^\s#]+)\s*(?:#.*)?$/gm)];
  if (matches.length !== 1) {
    throw new Error(`${DOCKERFILE}: expected exactly one ARG ZBASE_VERSION=<version>`);
  }

  const version = matches[0][1];
  if (!/^[0-9A-Za-z][0-9A-Za-z._-]*$/.test(version)) {
    throw new Error(`${DOCKERFILE}: invalid ZBASE_VERSION ${JSON.stringify(version)}`);
  }
  return version;
}

function parseVersion(value) {
  const match = value.trim().match(/^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$/);
  if (!match) throw new Error(`unsupported version: ${value}`);
  return {
    numeric: match.slice(1, 4).map(Number),
    prerelease: match[4] ?? "",
  };
}

export function compareVersions(left, right) {
  const a = parseVersion(left);
  const b = parseVersion(right);
  for (let index = 0; index < a.numeric.length; index += 1) {
    if (a.numeric[index] !== b.numeric[index]) return a.numeric[index] - b.numeric[index];
  }
  if (a.prerelease === b.prerelease) return 0;
  if (!a.prerelease) return 1;
  if (!b.prerelease) return -1;
  return a.prerelease.localeCompare(b.prerelease, undefined, { numeric: true });
}

export async function fetchBaseExtensions(version) {
  const tag = encodeURIComponent(`zbase/v${version}`);
  const source = `${ZIMAGE_RAW_URL}/refs/tags/${tag}/zbase/extensions.yaml`;
  const response = await fetch(source);
  if (!response.ok) {
    const detail = response.status === 404
      ? `zbase/v${version} does not contain zbase/extensions.yaml`
      : `${response.status} ${response.statusText}`;
    throw new Error(`failed to fetch extensions for ZBASE_VERSION=${version}: ${detail}\n${source}`);
  }
  return parseExtensions(await response.text(), source);
}

export function classifyExtensions(local, base) {
  const baseByName = new Map(base.map((entry) => [entry.name, entry]));
  const overrides = [];
  const downgrades = [];

  for (const entry of local) {
    const baseEntry = baseByName.get(entry.name);
    if (!baseEntry || entry.repo !== baseEntry.repo) {
      overrides.push(entry);
      continue;
    }
    const comparison = compareVersions(entry.tag, baseEntry.tag);
    if (comparison > 0) overrides.push(entry);
    if (comparison < 0) downgrades.push({ entry, baseEntry });
  }
  return { overrides, downgrades };
}

function shellQuote(value) {
  return `'${value.replaceAll("'", `'"'"'`)}'`;
}

export function renderGeneratedBlock(overrides) {
  if (overrides.length === 0) {
    return "";
  }

  const names = overrides.map(({ name }) => shellQuote(name)).join(" ");
  const lines = [
    GENERATED_START,
    "RUN --mount=from=composer:2.10,source=/usr/bin/composer,target=/usr/bin/composer \\",
    "    set -eux \\",
    "    && cd /app/w/extensions \\",
    `    && rm -rf ${names} \\`,
  ];
  for (const { name, repo, tag } of overrides) {
    lines.push(`    && git clone --depth=1 --branch ${shellQuote(tag)} ${shellQuote(repo)} ${shellQuote(name)} \\`);
  }
  lines.push(
    "    && cd /app/w \\",
    "    && composer update --minimal-changes --no-dev --no-scripts --optimize-autoloader \\",
    "    && chown www-data:www-data -R /app/*",
    GENERATED_END,
  );
  return lines.join("\n");
}

function generatedBlockPattern() {
  return /^# extensions-sync: generated; do not edit\r?\n[\s\S]*?^# \/extensions-sync: generated; do not edit$/gm;
}

export function readGeneratedBlock(dockerfile) {
  const matches = [...dockerfile.matchAll(generatedBlockPattern())];
  if (matches.length > 1) {
    throw new Error(`${DOCKERFILE}: expected at most one generated extension block`);
  }
  return matches[0]?.[0].replaceAll("\r\n", "\n").replace(/\n+$/, "") ?? "";
}

export function replaceGeneratedBlock(dockerfile, generatedBlock) {
  const currentBlock = readGeneratedBlock(dockerfile);
  if (!currentBlock) {
    return generatedBlock ? `${dockerfile.trimEnd()}\n\n${generatedBlock}\n` : dockerfile;
  }
  return dockerfile.replace(generatedBlockPattern(), generatedBlock).replace(/\n{2,}$/, "\n");
}
