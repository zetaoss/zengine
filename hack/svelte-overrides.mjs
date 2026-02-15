import { resolve } from "node:path";
import { readFileSync, writeFileSync } from "node:fs";

const ROOT = resolve(".");
const ROOT_PKG = resolve(ROOT, "package.json");
const PKG_A = resolve(ROOT, "mwz/skins/ZetaSkin/svelte/package.json");
const PKG_B = resolve(ROOT, "svelte/package.json");

const args = new Set(process.argv.slice(2));
const isFix = args.has("--fix");

function readJson(file) {
  return JSON.parse(readFileSync(file, "utf-8"));
}

function depsOf(pkg) {
  return {
    ...(pkg.dependencies ?? {}),
    ...(pkg.devDependencies ?? {}),
  };
}

function ensureObject(v) {
  return v && typeof v === "object" ? v : {};
}

const gray = (s) => `\x1b[90m${s}\x1b[0m`;
const red = (s) => `\x1b[91m${s}\x1b[0m`;
const green = (s) => `\x1b[92m${s}\x1b[0m`;
const yellow = (s) => `\x1b[93m${s}\x1b[0m`;

const rootPkg = readJson(ROOT_PKG);
const pkgA = readJson(PKG_A);
const pkgB = readJson(PKG_B);

const depsA = depsOf(pkgA);
const depsB = depsOf(pkgB);

const commonDeps = Object.keys(depsA)
  .filter((name) => name in depsB)
  .sort((a, b) => a.localeCompare(b));

const commonSet = new Set(commonDeps);

rootPkg.pnpm = ensureObject(rootPkg.pnpm);
rootPkg.pnpm.overrides = ensureObject(rootPkg.pnpm.overrides);

const overrides = rootPkg.pnpm.overrides;

const missing = commonDeps.filter((name) => !(name in overrides));
const versionMismatches = commonDeps.filter(
  (name) => depsA[name] !== depsB[name],
);

// ✅ 추가: 공통이 아닌데 overrides에 있는 항목
const extraneous = Object.keys(overrides)
  .filter((name) => !commonSet.has(name))
  .sort((a, b) => a.localeCompare(b));

console.log(gray("[svelte-overrides]"));
console.log(`root: ${ROOT_PKG}`);
console.log(`A   : ${PKG_A}`);
console.log(`B   : ${PKG_B}`);
console.log("");

if (!isFix) {
  if (
    missing.length === 0 &&
    extraneous.length === 0 &&
    versionMismatches.length === 0
  ) {
    console.log(
      green(
        `OK: root pnpm.overrides matches common deps (${commonDeps.length})`,
      ),
    );
    process.exit(0);
  }

  console.error(
    red("ERROR: root pnpm.overrides is out of sync with common deps"),
  );
  console.error("");

  if (missing.length > 0) {
    console.error(red(`Missing in overrides: ${missing.length}`));
    for (const name of missing) {
      console.error(`- ${name}  (A: ${depsA[name]})  (B: ${depsB[name]})`);
    }
    console.error("");
  }

  if (extraneous.length > 0) {
    console.error(
      red(`Extraneous in overrides (not common): ${extraneous.length}`),
    );
    for (const name of extraneous) {
      console.error(`- ${name}  (override: ${overrides[name]})`);
    }
    console.error("");
  }

  if (versionMismatches.length > 0) {
    console.error(
      red(`Version mismatches between A and B: ${versionMismatches.length}`),
    );
    for (const name of versionMismatches) {
      console.error(`- ${name}  (A: ${depsA[name]})  (B: ${depsB[name]})`);
    }
    console.error("");
  }

  console.error("Fix: edit root package.json -> pnpm.overrides, or run:");
  console.error("  node hack/svelte-overrides.mjs --fix");
  process.exit(1);
}

// --fix: make overrides exactly match common deps (A's spec), and remove extraneous
console.log(
  yellow(
    `--fix: adding ${missing.length} missing, removing ${extraneous.length} extraneous in root pnpm.overrides...`,
  ),
);

if (versionMismatches.length > 0) {
  console.log(
    yellow(
      `Note: ${versionMismatches.length} version mismatches detected between A and B (no changes made).`,
    ),
  );
  for (const name of versionMismatches) {
    console.log(`- ${name}  (A: ${depsA[name]})  (B: ${depsB[name]})`);
  }
  console.log("");
}

// add missing (A spec)
for (const name of missing) {
  overrides[name] = depsA[name];
}

// remove extraneous
for (const name of extraneous) {
  delete overrides[name];
}

// sort for stable diff
const sortedOverrides = {};
for (const k of Object.keys(overrides).sort((a, b) => a.localeCompare(b))) {
  sortedOverrides[k] = overrides[k];
}
rootPkg.pnpm.overrides = sortedOverrides;

writeFileSync(ROOT_PKG, JSON.stringify(rootPkg, null, 2) + "\n", "utf-8");

console.log(
  green(
    `Updated root pnpm.overrides. Added: ${missing.length}, Removed: ${extraneous.length}`,
  ),
);
console.log(gray("Next: run pnpm install to update lockfile."));
process.exit(0);
