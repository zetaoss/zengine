import {
  cpSync,
  existsSync,
  lstatSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  renameSync,
  rmSync,
} from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const CONFIG = resolve(ROOT, "hack/extensions.yaml");
const EXTENSIONS_DIR = resolve(process.env.EXTENSIONS_DIR ?? resolve(ROOT, "w/extensions"));

function parseExtensions(source) {
  const entries = [];
  let entry;

  for (const [index, rawLine] of source.split("\n").entries()) {
    const lineNumber = index + 1;
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) continue;

    const field = line.match(/^(?:- )?(name|repo|tag):\s*(.*?)\s*$/);
    if (field) {
      const [, key, value] = field;
      if (key === "name") {
        if (entry) entries.push(entry);
        entry = { name: value };
      } else {
        if (!entry) throw new Error(`${CONFIG}:${lineNumber}: ${key} precedes name`);
        entry[key] = value;
      }
      continue;
    }

    throw new Error(`${CONFIG}:${lineNumber}: unsupported YAML syntax`);
  }
  if (entry) entries.push(entry);

  const names = new Set();
  for (const extension of entries) {
    for (const key of ["name", "repo", "tag"]) {
      if (!extension[key]) throw new Error(`${CONFIG}: missing ${key}`);
    }
    if (!/^[A-Za-z0-9._-]+$/.test(extension.name)) {
      throw new Error(`${CONFIG}: invalid extension name ${JSON.stringify(extension.name)}`);
    }
    if (names.has(extension.name)) throw new Error(`${CONFIG}: duplicate extension ${extension.name}`);
    names.add(extension.name);
  }
  return entries;
}

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

const entries = parseExtensions(readFileSync(CONFIG, "utf8"));
mkdirSync(EXTENSIONS_DIR, { recursive: true });

for (const entry of entries) {
  const target = extensionPath(entry.name);
  if (existsSync(target)) {
    console.log(`Removing ${entry.name}`);
    rmSync(target, { recursive: true, force: true });
  }

  console.log(`Installing ${entry.name} (${entry.tag})`);
  run("git", ["clone", "--depth=1", "--branch", entry.tag, entry.repo, target]);
  if (applyOverrides(entry)) {
    console.log(`Applying local overrides for ${entry.name}`);
  }
}

console.log(`✅  Installed ${entries.length} extensions`);
