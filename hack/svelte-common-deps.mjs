import { readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";

const ROOT = resolve(".");
const MAIN_PKG = resolve(ROOT, "svelte/package.json");
const SKIN_PKG = resolve(ROOT, "mwz/skins/ZetaSkin/svelte/package.json");

const args = new Set(process.argv.slice(2));
const isFix = args.has("--fix");

function readJson(file) {
  return JSON.parse(readFileSync(file, "utf-8"));
}

function writeJson(file, value) {
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`, "utf-8");
}

function depsOf(pkg) {
  return {
    ...(pkg.dependencies ?? {}),
    ...(pkg.devDependencies ?? {}),
  };
}

function setDependencySpec(pkg, name, spec) {
  if (pkg.dependencies && name in pkg.dependencies) {
    pkg.dependencies[name] = spec;
    return;
  }
  if (pkg.devDependencies && name in pkg.devDependencies) {
    pkg.devDependencies[name] = spec;
  }
}

const gray = (s) => `\x1b[90m${s}\x1b[0m`;
const red = (s) => `\x1b[91m${s}\x1b[0m`;
const green = (s) => `\x1b[92m${s}\x1b[0m`;
const yellow = (s) => `\x1b[93m${s}\x1b[0m`;

const mainPkg = readJson(MAIN_PKG);
const skinPkg = readJson(SKIN_PKG);

const mainDeps = depsOf(mainPkg);
const skinDeps = depsOf(skinPkg);

const commonDeps = Object.keys(mainDeps)
  .filter((name) => name in skinDeps)
  .sort((a, b) => a.localeCompare(b));

const mismatches = commonDeps.filter(
  (name) => mainDeps[name] !== skinDeps[name],
);

console.log(gray("[svelte-common-deps]"));
console.log(`main: ${MAIN_PKG}`);
console.log(`skin: ${SKIN_PKG}`);
console.log("");

if (mismatches.length === 0) {
  console.log(green(`OK: common dependency specs match (${commonDeps.length})`));
  process.exit(0);
}

if (!isFix) {
  console.error(
    red(`ERROR: common dependency specs differ (${mismatches.length})`),
  );
  console.error("");
  for (const name of mismatches) {
    console.error(
      `- ${name}  (main: ${mainDeps[name]})  (skin: ${skinDeps[name]})`,
    );
  }
  console.error("");
  console.error("Fix: align the package.json files, or run:");
  console.error("  node hack/svelte-common-deps.mjs --fix");
  process.exit(1);
}

console.log(
  yellow(
    `--fix: aligning ${mismatches.length} skin dependency specs with main svelte...`,
  ),
);

for (const name of mismatches) {
  setDependencySpec(skinPkg, name, mainDeps[name]);
}

writeJson(SKIN_PKG, skinPkg);

console.log(green(`Updated ${SKIN_PKG}`));
console.log(
  gray("Next: run pnpm -C mwz/skins/ZetaSkin/svelte install to update lockfile."),
);
process.exit(0);
