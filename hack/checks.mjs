import { createHash } from "node:crypto";
import { existsSync, mkdirSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import { spawnSync } from "node:child_process";

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const CACHE_DIR = "/tmp/make-checks";
const IGNORED_DIRECTORIES = new Set([".git", ".svelte-kit", "bin", "dist", "node_modules", "tmp", "vendor"]);

function run(command, args, { cwd = ROOT, fix } = {}) {
  console.log(`➡️  ${cwd === ROOT ? "" : `${cwd.slice(ROOT.length + 1)}: `}${command} ${args.join(" ")}`);
  const result = spawnSync(command, args, { cwd, stdio: "inherit" });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    if (fix) console.log(`\n💡 This might be fixed with:\n\n   ${fix}\n`);
    process.exit(result.status ?? 1);
  }
}

function runMake(target) {
  run("make", [target]);
}

function checkExtension(path) {
  run("composer", ["install", "--no-interaction", "--prefer-dist"], { cwd: resolve(ROOT, path) });
  run("composer", ["test"], { cwd: resolve(ROOT, path), fix: `cd ${path} && composer fix` });
}

function checkMainSvelte() {
  run("node", ["hack/svelte-common-deps.mjs"]);
  checkSvelteProject("svelte");
}

function checkSvelteProject(path) {
  const cwd = resolve(ROOT, path);
  const fix = `pnpm -C ${path} format:fix`;
  run("pnpm", ["install", "--frozen-lockfile"], { cwd });
  run("pnpm", ["peers", "check"], { cwd, fix });
  run("pnpm", ["lint"], { cwd, fix });
  run("pnpm", ["build"], { cwd, fix });
}

function checkGoapp() {
  const cwd = resolve(ROOT, "goapp");
  run("make", ["lint"], { cwd });
  run("go", ["test", "./..."], { cwd });
  run("go", ["build", "./..."], { cwd });
}

const cachedChecks = [
  { name: "check-extension", paths: ["mwz/extensions/ZetaExtension"], action: () => checkExtension("mwz/extensions/ZetaExtension") },
  { name: "check-skin", paths: ["mwz/skins/ZetaSkin"], action: () => checkExtension("mwz/skins/ZetaSkin") },
  { name: "check-main-svelte", paths: ["svelte", "hack/svelte-common-deps.mjs", "pnpm-lock.yaml", "package.json"], action: checkMainSvelte },
  { name: "check-skin-svelte", paths: ["mwz/skins/ZetaSkin/svelte", "svelte/src/shared", "pnpm-lock.yaml", "package.json"], action: () => checkSvelteProject("mwz/skins/ZetaSkin/svelte") },
  { name: "check-goapp", paths: ["goapp", ".golangci.yml"], action: checkGoapp },
];

const checksByName = new Map(cachedChecks.map((check) => [check.name, check]));

function toolVersion(command, args) {
  const result = spawnSync(command, args, { encoding: "utf8" });
  return result.status === 0 ? result.stdout.trim() : "";
}

function hashPath(hash, path) {
  if (!existsSync(path)) return;
  const stat = statSync(path);
  if (stat.isFile()) {
    if (!path.endsWith(".md")) {
      hash.update(`${path}\0`);
      hash.update(readFileSync(path));
    }
    return;
  }
  if (!stat.isDirectory()) return;

  for (const entry of readdirSync(path, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
    if (entry.isDirectory() && IGNORED_DIRECTORIES.has(entry.name)) continue;
    hashPath(hash, resolve(path, entry.name));
  }
}

function cacheKey(check) {
  const hash = createHash("sha256");
  hash.update(`check=${check.name}\n`);
  for (const [command, args] of [["node", ["-v"]], ["pnpm", ["-v"]], ["php", ["-v"]], ["go", ["version"]]]) {
    hash.update(`${command}=${toolVersion(command, args)}\n`);
  }
  hashPath(hash, resolve(ROOT, "Makefile"));
  hashPath(hash, resolve(ROOT, "hack/checks.mjs"));
  for (const path of check.paths) hashPath(hash, resolve(ROOT, path));
  return hash.digest("hex");
}

function runCached(check) {
  const hashFile = resolve(CACHE_DIR, `${check.name}.hash`);
  const key = cacheKey(check);
  if (existsSync(hashFile) && readFileSync(hashFile, "utf8") === key) {
    console.log(`⏭️  ${check.name}: no changes, skip`);
    return;
  }
  console.log(`➡️  ${check.name}: running`);
  check.action();
  mkdirSync(CACHE_DIR, { recursive: true });
  writeFileSync(hashFile, key);
}

function runAll(useCache) {
  runMake("extensions-check");
  for (const check of cachedChecks) {
    if (useCache) runCached(check);
    else check.action();
  }
  console.log(`✅  All checks passed${useCache ? "" : " (no cache)"}`);
}

function clearCache() {
  console.log(`🧹 clear cache: ${CACHE_DIR}`);
  rmSync(CACHE_DIR, { recursive: true, force: true });
}

const [target = "checks", option] = process.argv.slice(2);
if (target === "checks") {
  runAll(option !== "--no-cache");
} else if (target === "clear") {
  clearCache();
} else if (target === "check-php") {
  checksByName.get("check-extension").action();
  checksByName.get("check-skin").action();
} else if (target === "check-svelte") {
  checksByName.get("check-main-svelte").action();
  checksByName.get("check-skin-svelte").action();
} else if (checksByName.has(target)) {
  checksByName.get(target).action();
} else {
  throw new Error(`Unknown check target: ${target}`);
}
