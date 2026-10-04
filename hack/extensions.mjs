// Extra (non-bundled) MediaWiki extensions from mwz/extensions.yaml (see the header of that file).
//
//   node hack/extensions.mjs [install]      clone the extensions into EXTENSIONS_DIR
//   node hack/extensions.mjs settings [out] write ExtraExtensionSettings.php (stdout when out is omitted)
//
// Requires hack/node_modules (pnpm -C hack install).
import {
  cpSync,
  existsSync,
  lstatSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import { parse } from "yaml";

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const MW_DIR = resolve(ROOT, "mwz");
const CONFIG = resolve(MW_DIR, "extensions.yaml");
const COMMIT_SHA = /^[0-9a-f]{40}$/;
const NAME = /^[A-Za-z0-9._-]+$/;
const FIELDS = new Set(["name", "repo", "tag"]);
const EXTENSIONS_DIR = resolve(process.env.EXTENSIONS_DIR ?? resolve(ROOT, "w/extensions"));

function fail(message) {
  throw new Error(`${CONFIG}: ${message}`);
}

function loadExtensions() {
  const entries = parse(readFileSync(CONFIG, "utf8"));
  if (!Array.isArray(entries)) fail("expected a list of extensions");

  const names = new Set();
  for (const entry of entries) {
    const where = `extension ${entry?.name ?? "?"}`;
    if (typeof entry !== "object" || entry === null) fail(`${where}: expected a mapping`);
    for (const key of Object.keys(entry)) {
      if (!FIELDS.has(key)) fail(`${where}: unknown field ${key}`);
    }
    if (typeof entry.name !== "string" || !NAME.test(entry.name)) {
      fail(`${where}: invalid name ${JSON.stringify(entry.name)}`);
    }
    if (names.has(entry.name)) fail(`duplicate extension ${entry.name}`);
    names.add(entry.name);

    if (typeof entry.repo !== "string" || typeof entry.tag !== "string") fail(`${where}: repo and tag are required`);
  }
  return entries;
}

// ---- install ----

function run(command, args, cwd = ROOT) {
  const result = spawnSync(command, args, { cwd, stdio: "inherit" });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}

function extensionPath(name) {
  const target = resolve(EXTENSIONS_DIR, name);
  if (!target.startsWith(`${EXTENSIONS_DIR}/`)) {
    throw new Error(`Refusing to access extension outside ${EXTENSIONS_DIR}: ${name}`);
  }
  return target;
}

function applyOverrides(extension) {
  const source = resolve(ROOT, "mwz/extensions", extension.name);
  if (!existsSync(source)) return false;

  const target = extensionPath(extension.name);
  copyOverrides(source, target);
  return true;
}

function pathExists(path) {
  try {
    lstatSync(path);
    return true;
  } catch (error) {
    if (error.code === "ENOENT") return false;
    throw error;
  }
}

function copyOverrides(source, target) {
  mkdirSync(target, { recursive: true });
  for (const item of readdirSync(source, { withFileTypes: true })) {
    const sourcePath = resolve(source, item.name);
    const targetPath = resolve(target, item.name);
    if (item.isDirectory()) {
      if (pathExists(targetPath) && !lstatSync(targetPath).isDirectory()) {
        moveToBackup(targetPath);
      }
      copyOverrides(sourcePath, targetPath);
      continue;
    }

    if (pathExists(targetPath)) moveToBackup(targetPath);
    cpSync(sourcePath, targetPath, { recursive: true, force: true });
  }
}

function moveToBackup(path) {
  const backupPath = `${path}.bak`;
  if (pathExists(backupPath)) {
    throw new Error(`Backup already exists, refusing to overwrite: ${backupPath}`);
  }
  console.log(`Backing up ${path} -> ${backupPath}`);
  renameSync(path, backupPath);
}

function install(entries) {
  const targets = entries;
  mkdirSync(EXTENSIONS_DIR, { recursive: true });

  for (const entry of targets) {
    const target = extensionPath(entry.name);
    if (existsSync(target)) {
      console.log(`Removing ${entry.name}`);
      rmSync(target, { recursive: true, force: true });
    }

    console.log(`Installing ${entry.name} (${entry.tag})`);
    if (COMMIT_SHA.test(entry.tag)) {
      // A branch/tag clone cannot target a commit, so fetch the single commit instead.
      run("git", ["init", "--quiet", target]);
      run("git", ["-C", target, "remote", "add", "origin", entry.repo]);
      run("git", ["-C", target, "fetch", "--quiet", "--depth=1", "origin", entry.tag]);
      run("git", ["-C", target, "checkout", "--quiet", "FETCH_HEAD"]);
    } else {
      run("git", ["clone", "--depth=1", "--branch", entry.tag, entry.repo, target]);
    }
    if (applyOverrides(entry)) {
      console.log(`Applying local overrides for ${entry.name}`);
    }
  }

  console.log(`✅  Installed ${targets.length} extensions`);
}

// ---- settings ----

function phpString(value) {
  return `'${value.replaceAll("\\", "\\\\").replaceAll("'", "\\'")}'`;
}

function settings(entries) {
  const lines = [
    "<?php",
    "",
    "// ExtraExtensionSettings.php: generated from mwz/extensions.yaml by hack/extensions.mjs. Do not edit.",
  ];
  for (const entry of entries) {
    lines.push("", `// ${entry.name}`);
    lines.push(`wfLoadExtension(${phpString(entry.name)});`);
  }
  return `${lines.join("\n")}\n`;
}

// ---- main ----

const [command = "install", out] = process.argv.slice(2);
const entries = loadExtensions();
if (command === "install") {
  install(entries);
} else if (command === "settings") {
  const php = settings(entries);
  if (out) writeFileSync(out, php);
  else process.stdout.write(php);
} else {
  throw new Error(`unknown command ${command}; expected install or settings`);
}
