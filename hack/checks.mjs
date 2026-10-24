import { createHash } from "node:crypto";
import { existsSync, mkdirSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import { spawn, spawnSync } from "node:child_process";
import { availableParallelism } from "node:os";

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const CACHE_DIR = process.env.CHECKS_CACHE_DIR || resolve(ROOT, ".checks");
const IGNORED_DIRECTORIES = new Set([".checks", ".git", ".svelte-kit", "bin", "dist", "node_modules", "tmp", "vendor"]);

class CheckFailed extends Error {
  constructor(status, fix) {
    super(`exit ${status}`);
    this.status = status;
    this.fix = fix;
  }
}

// runner returns the run function of one check. Its output goes to out: the terminal for a single check, or a
// buffer printed when the check ends, so checks running in parallel do not interleave.
function runner(out) {
  return (command, args, { cwd = ROOT, fix } = {}) =>
    new Promise((resolvePromise, reject) => {
      out.write(`➡️  ${cwd === ROOT ? "" : `${cwd.slice(ROOT.length + 1)}: `}${command} ${args.join(" ")}\n`);
      const child = spawn(command, args, { cwd, stdio: out === process.stdout ? "inherit" : ["ignore", "pipe", "pipe"] });
      child.stdout?.on("data", (d) => out.write(d));
      child.stderr?.on("data", (d) => out.write(d));
      child.on("error", reject);
      child.on("close", (status) => (status === 0 ? resolvePromise() : reject(new CheckFailed(status ?? 1, fix))));
    });
}

async function checkExtension(run, path) {
  await run("composer", ["install", "--no-interaction", "--prefer-dist"], { cwd: resolve(ROOT, path) });
  await run("composer", ["test"], { cwd: resolve(ROOT, path), fix: `cd ${path} && composer fix` });
}

async function checkMainSvelte(run) {
  await run("node", ["hack/svelte-common-deps.mjs"]);
  await run("node", ["hack/check-shared.mjs"]);
  await checkSvelteProject(run, "svelte", "check");
}

// The type check stands in for a build: the docker-build job builds both apps for the image anyway.
async function checkSvelteProject(run, path, typeCheck) {
  const cwd = resolve(ROOT, path);
  const fix = `pnpm -C ${path} format:fix`;
  await run("pnpm", ["install", "--frozen-lockfile"], { cwd });
  await run("pnpm", ["peers", "check"], { cwd, fix });
  await run("pnpm", ["lint"], { cwd, fix });
  await run("pnpm", [typeCheck], { cwd });
}

// golangci-lint type-checks every package and the docker-build job builds the binaries, so there is no go build.
async function checkGoapp(run) {
  const cwd = resolve(ROOT, "goapp");
  await run("make", ["lint"], { cwd });
  await run("go", ["test", "./..."], { cwd });
}

// mwz/extensions.yaml is valid and mwz/extensions.lock matches it.
async function checkMwExtensions(run) {
  await run("pnpm", ["-C", "hack", "install", "--frozen-lockfile"]);
  await run("node", ["hack/extensions.mjs", "check"], { fix: "make extensions-lock" });
}

const cachedChecks = [
  { name: "check-mw-extensions", paths: ["mwz/extensions.yaml", "mwz/extensions.lock", "hack/extensions.mjs", "hack/package.json", "hack/pnpm-lock.yaml"], action: checkMwExtensions },
  { name: "check-extension", paths: ["mwz/extensions/ZetaExtension"], action: (run) => checkExtension(run, "mwz/extensions/ZetaExtension") },
  { name: "check-skin", paths: ["mwz/skins/ZetaSkin"], action: (run) => checkExtension(run, "mwz/skins/ZetaSkin") },
  { name: "check-main-svelte", paths: ["svelte", "mwz/skins/ZetaSkin/svelte/src", "hack/svelte-common-deps.mjs", "hack/check-shared.mjs", "pnpm-lock.yaml", "package.json"], action: checkMainSvelte },
  { name: "check-skin-svelte", paths: ["mwz/skins/ZetaSkin/svelte", "svelte/src/shared", "pnpm-lock.yaml", "package.json"], action: (run) => checkSvelteProject(run, "mwz/skins/ZetaSkin/svelte", "type-check") },
  { name: "check-goapp", paths: ["goapp", ".golangci.yml"], action: checkGoapp },
];

const checksByName = new Map(cachedChecks.map((check) => [check.name, check]));

function toolVersion(command, args) {
  const result = spawnSync(command, args, { encoding: "utf8" });
  return result.status === 0 ? result.stdout.trim() : "";
}

const toolVersions = [["node", ["-v"]], ["pnpm", ["-v"]], ["php", ["-v"]], ["go", ["version"]]]
  .map(([command, args]) => `${command}=${toolVersion(command, args)}\n`)
  .join("");

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
  hash.update(toolVersions);
  hashPath(hash, resolve(ROOT, "Makefile"));
  hashPath(hash, resolve(ROOT, "hack/checks.mjs"));
  for (const path of check.paths) hashPath(hash, resolve(ROOT, path));
  return hash.digest("hex");
}

// Buffer collects a check's output while it runs in parallel with others.
class Buffer {
  chunks = [];
  write(d) {
    this.chunks.push(typeof d === "string" ? d : d.toString());
  }
  toString() {
    return this.chunks.join("");
  }
}

// runOne runs a check unless its inputs are unchanged since it last passed, and returns its result.
async function runOne(check, useCache, out) {
  const hashFile = resolve(CACHE_DIR, `${check.name}.hash`);
  const key = useCache ? cacheKey(check) : "";
  if (useCache && existsSync(hashFile) && readFileSync(hashFile, "utf8") === key) {
    return { check, skipped: true };
  }
  const started = Date.now();
  try {
    await check.action(runner(out));
  } catch (err) {
    return { check, err, seconds: (Date.now() - started) / 1000 };
  }
  if (useCache) {
    mkdirSync(CACHE_DIR, { recursive: true });
    writeFileSync(hashFile, key);
  }
  return { check, seconds: (Date.now() - started) / 1000 };
}

// runAll runs the checks in parallel, CHECKS_JOBS at a time (default: the CPU count), printing each one's
// output when it ends, and fails after all of them if any failed.
async function runAll(useCache) {
  const jobs = Number(process.env.CHECKS_JOBS) || availableParallelism();
  const queue = [...cachedChecks];
  const failed = [];
  const worker = async () => {
    for (let check = queue.shift(); check; check = queue.shift()) {
      const out = new Buffer();
      const result = await runOne(check, useCache, out);
      if (result.skipped) {
        console.log(`⏭️  ${check.name}: no changes, skip`);
        continue;
      }
      const mark = result.err ? "❌" : "✔️ ";
      console.log(`\n${mark} ${check.name} (${result.seconds.toFixed(1)}s)\n${out}`);
      if (result.err) failed.push(result);
    }
  };
  await Promise.all(Array.from({ length: Math.min(jobs, cachedChecks.length) }, worker));
  if (failed.length) {
    for (const { check, err } of failed) {
      console.log(`❌  ${check.name} failed: ${err.message}`);
      if (err.fix) console.log(`\n💡 This might be fixed with:\n\n   ${err.fix}\n`);
    }
    process.exit(1);
  }
  console.log(`✅  All checks passed${useCache ? "" : " (no cache)"}`);
}

// runDirect runs checks one after another with their output on the terminal, for make check-<name>.
async function runDirect(...checks) {
  const run = runner(process.stdout);
  for (const check of checks) {
    try {
      await check.action(run);
    } catch (err) {
      if (err.fix) console.log(`\n💡 This might be fixed with:\n\n   ${err.fix}\n`);
      process.exit(err.status ?? 1);
    }
  }
}

function clearCache() {
  console.log(`🧹 clear cache: ${CACHE_DIR}`);
  rmSync(CACHE_DIR, { recursive: true, force: true });
}

const [target = "checks", option] = process.argv.slice(2);
if (target === "checks") {
  await runAll(option !== "--no-cache");
} else if (target === "clear") {
  clearCache();
} else if (target === "check-php") {
  await runDirect(checksByName.get("check-extension"), checksByName.get("check-skin"));
} else if (target === "check-svelte") {
  await runDirect(checksByName.get("check-main-svelte"), checksByName.get("check-skin-svelte"));
} else if (checksByName.has(target)) {
  await runDirect(checksByName.get(target));
} else {
  throw new Error(`Unknown check target: ${target}`);
}
