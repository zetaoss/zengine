// Fails when a file in svelte/src/shared is not imported by both the main and the skin Svelte app.
import { existsSync, readdirSync, readFileSync, realpathSync, statSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const SHARED = join(ROOT, "svelte/src/shared");
const APPS = {
  main: { src: join(ROOT, "svelte/src"), lib: join(ROOT, "svelte/src/lib") },
  skin: { src: join(ROOT, "mwz/skins/ZetaSkin/svelte/src"), lib: join(ROOT, "mwz/skins/ZetaSkin/svelte/src/lib") },
};
// Written by CI before the build.
const EXEMPT = new Set(["version.json"]);
const EXTENSIONS = ["", ".ts", ".js", ".svelte", ".css", ".json", "/index.ts", "/index.js"];
const SPECIFIERS = [
  /\bfrom\s*['"]([^'"]+)['"]/g,
  /\bimport\s*['"]([^'"]+)['"]/g,
  /\bimport\(\s*['"]([^'"]+)['"]\s*\)/g,
  /@import\s+(?:url\()?['"]([^'"]+)['"]/g,
  /@reference\s+['"]([^'"]+)['"]/g,
];

const inShared = (file) => file.startsWith(`${SHARED}/`);
const isTest = (file) => /\.(test|spec)\.ts$/.test(file);

function walk(dir, files = []) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isSymbolicLink()) continue;
    if (entry.isDirectory()) walk(path, files);
    else if (/\.(ts|js|svelte|css|json|html)$/.test(entry.name)) files.push(path);
  }
  return files;
}

function resolveSpecifier(specifier, from, app) {
  const spec = specifier.split("?")[0];
  let base;
  if (spec === "$shared" || spec.startsWith("$shared/")) base = join(SHARED, spec.slice("$shared".length));
  else if (spec === "$lib" || spec.startsWith("$lib/")) base = join(app.lib, spec.slice("$lib".length));
  else if (spec.startsWith(".")) base = resolve(dirname(from), spec);
  else return null;
  for (const ext of EXTENSIONS) {
    const path = base + ext;
    if (existsSync(path) && statSync(path).isFile()) return realpathSync(path);
  }
  return null;
}

function imports(file, app) {
  const text = readFileSync(file, "utf8");
  const found = new Set();
  for (const re of SPECIFIERS) {
    for (const match of text.matchAll(re)) {
      const path = resolveSpecifier(match[1], file, app);
      if (path) found.add(path);
    }
  }
  return found;
}

function sharedFilesUsedBy(app) {
  // A test import does not make a shared file used by the app.
  const roots = walk(app.src).filter((file) => !inShared(file) && !isTest(file));
  const used = new Set();
  const stack = roots.flatMap((file) => [...imports(file, app)].filter(inShared));
  while (stack.length > 0) {
    const file = stack.pop();
    if (used.has(file)) continue;
    used.add(file);
    for (const dep of imports(file, app)) if (inShared(dep)) stack.push(dep);
  }
  return used;
}

const used = Object.fromEntries(Object.entries(APPS).map(([name, app]) => [name, sharedFilesUsedBy(app)]));
const problems = [];
for (const file of walk(SHARED)) {
  const path = relative(SHARED, file);
  if (EXEMPT.has(path) || isTest(file)) continue;
  const users = Object.keys(APPS).filter((name) => used[name].has(file));
  if (users.length === 2) continue;
  problems.push(users.length === 0 ? `${path}: not used; delete it` : `${path}: used only by ${users[0]}; move it to ${relative(ROOT, APPS[users[0]].lib)}/`);
}

if (problems.length > 0) {
  console.error(`❌ svelte/src/shared must hold only code used by both Svelte apps:\n${problems.map((p) => `   ${p}`).join("\n")}`);
  process.exit(1);
}
console.log("✅ svelte/src/shared: every file is used by both Svelte apps");
