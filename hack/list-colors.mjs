#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT_DIR = path.resolve(SCRIPT_DIR, "..");
const SEARCH_DIRS = [
  path.join(ROOT_DIR, "svelte"),
  path.join(ROOT_DIR, "mwz/skins/ZetaSkin/svelte"),
];
const query = process.argv.slice(2).join(" ").trim().toLowerCase();

const regex =
  /\b(?:bg|text|border|ring|outline|fill|stroke|from|via|to|decoration|accent|caret|shadow)-(?:\[[^\]\s'"]+\]|\(--[A-Za-z0-9_-]+\)|[a-z0-9]+(?:-[a-z0-9]+)*)(?:\/\d{1,3})?!?/g;

const extensions = new Set([
  ".js",
  ".ts",
  ".jsx",
  ".tsx",
  ".vue",
  ".svelte",
  ".html",
  ".css",
]);

const ignoredDirectories = new Set([
  ".git",
  ".svelte-kit",
  "node_modules",
  "dist",
  "build",
]);

function listFiles(dir) {
  const files = [];

  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      if (!ignoredDirectories.has(entry.name)) {
        files.push(...listFiles(path.join(dir, entry.name)));
      }
      continue;
    }

    if (
      entry.isFile() &&
      !entry.name.endsWith(".d.ts") &&
      extensions.has(path.extname(entry.name))
    ) {
      files.push(path.join(dir, entry.name));
    }
  }

  return files;
}

function isArbitraryColor(cls) {
  if (/-\(--[A-Za-z0-9_-]+\)$/.test(cls)) {
    return true;
  }

  const value = cls.match(/-\[([^\]]+)\]/)?.[1];
  if (!value) {
    return true;
  }

  return (
    value.startsWith("#") ||
    value.startsWith("rgb") ||
    value.startsWith("hsl") ||
    value.startsWith("oklch") ||
    value.startsWith("color-mix") ||
    value.startsWith("var(--")
  );
}

const nonColorTextValues = new Set([
  "align",
  "balance",
  "center",
  "clip",
  "decoration",
  "decoration-line",
  "ellipsis",
  "end",
  "justify",
  "left",
  "nowrap",
  "pretty",
  "rendering",
  "right",
  "start",
  "wrap",
]);

function isTextSize(cls) {
  const value = cls.match(/^text-(.+)$/)?.[1];
  if (!value) {
    return false;
  }

  if (/^(?:xs|sm|lg|xl|[2-9]xl)$/.test(value)) {
    return true;
  }

  const arbitraryValue = value.match(/^\[([^\]]+)\]$/)?.[1];
  if (!arbitraryValue) {
    return false;
  }

  return (
    /^-?\d*\.?\d+(?:px|r?em|lh|ch|ex|cap|ic|vw|vh|vmin|vmax|svw|svh|lvw|lvh|dvw|dvh|cqw|cqh|cqi|cqb|cqmin|cqmax|%)$/i.test(
      arbitraryValue,
    ) ||
    /^(?:calc|min|max|clamp)\(/i.test(arbitraryValue)
  );
}

function isNonColorUtility(cls) {
  const textValue = cls.match(/^text-(.+)$/)?.[1];
  if (textValue) {
    return (
      isTextSize(cls) ||
      nonColorTextValues.has(textValue) ||
      textValue.startsWith("decoration-") ||
      textValue.startsWith("opacity-")
    );
  }

  if (
    /^bg-(?:bottom|center|clip|clip-.+|contain|cover|fixed|left|local|no-repeat|repeat|repeat-.+|right|scroll|top)$/.test(
      cls,
    )
  ) {
    return true;
  }

  return (
    /^(?:border|ring|outline|shadow)-(?:0|1|2|4|8|x|y|s|e|t|r|b|l|solid|dashed|dotted|double|hidden|none|sm|md|lg|xl|2xl|inner)$/.test(
      cls,
    ) ||
    /^(?:border|outline)-(?:[trblxyse]-)?(?:0|1|2|4|8)$/.test(cls) ||
    /^border-(?:bottom|bottom-color|box|collapse|color|color-.+|radius|width)$/.test(
      cls,
    ) ||
    /^outline-offset/.test(cls) ||
    /^ring-(?:0|1|2|3|4|8|inset|offset|offset-.+)$/.test(cls) ||
    /^shadow-color-.+/.test(cls)
  );
}

const files = SEARCH_DIRS.flatMap((dir) => listFiles(dir));
const rows = [];

for (const file of files) {
  const text = fs.readFileSync(file, "utf8");
  const lines = text.split("\n");

  for (const [index, line] of lines.entries()) {
    const classes = new Map();

    for (const match of line.matchAll(regex)) {
      const cls = match[0].replace(/\/\d{1,3}/, "").replace(/!$/, "");
      if (isNonColorUtility(cls)) {
        continue;
      }

      if (!isArbitraryColor(cls)) {
        continue;
      }

      const column = (match.index ?? 0) + match[0].length + 1;
      classes.set(`${cls}:${column}`, { cls, column });
    }

    for (const { cls, column } of classes.values()) {
      rows.push({
        cls,
        file: path.relative(ROOT_DIR, file),
        line: index + 1,
        column,
      });
    }
  }
}

function compareFile(a, b) {
  return (
    a.file.localeCompare(b.file) ||
    a.line - b.line ||
    a.column - b.column ||
    a.cls.localeCompare(b.cls)
  );
}

rows.sort(compareFile);

if (query) {
  const filteredRows = rows.filter((row) =>
    row.cls.toLowerCase().includes(query),
  );
  const fileWidth = Math.max(
    "FILE".length,
    ...filteredRows.map((row) => `${row.file}:${row.line}:${row.column}`.length),
  );
  const rowNumberWidth = Math.max(
    "#".length,
    filteredRows.length.toString().length,
  );

  console.log(
    `${"#".padStart(rowNumberWidth)}  ${"FILE".padEnd(fileWidth)}  CLASS`,
  );
  for (const [index, row] of filteredRows.entries()) {
    const fileLink = formatFileLink(row, fileWidth);
    console.log(
      `${(index + 1).toString().padStart(rowNumberWidth)}  ${fileLink}  ${row.cls}`,
    );
  }
  process.exit(0);
}

const stats = new Map();
for (const row of rows) {
  const match = row.cls.match(/^(.+)-(\d{2,3})$/);
  const name = match?.[1] ?? row.cls;
  const shade = match?.[2] ?? "base";
  const shades = stats.get(name) ?? new Map();
  shades.set(shade, (shades.get(shade) ?? 0) + 1);
  stats.set(name, shades);
}

const statRows = [...stats.entries()]
  .map(([name, shades]) => ({
    name,
    shades,
    total: [...shades.values()].reduce((sum, count) => sum + count, 0),
  }))
  .sort((a, b) => a.name.localeCompare(b.name));

const statShades = [
  ...new Set(statRows.flatMap((row) => [...row.shades.keys()])),
].sort((a, b) => {
  if (a === "base") {
    return -1;
  }

  if (b === "base") {
    return 1;
  }

  return Number(a) - Number(b);
});

const classWidth = Math.max(
  "CLASS".length,
  ...statRows.map((row) => row.name.length),
  ...rows.map((row) => row.cls.length),
);
const shadeWidth = Math.max(
  3,
  ...statRows.flatMap((row) =>
    statShades.map((shade) => (row.shades.get(shade) ?? ".").toString().length),
  ),
);

function formatFileLink(row, width) {
  return `${row.file}:${row.line}:${row.column}`.padEnd(width);
}

console.log(
  [
    "CLASS".padEnd(classWidth),
    ...statShades.map((shade) => shade.padStart(shadeWidth)),
  ].join("  "),
);
for (const row of statRows) {
  console.log(
    [
      row.name.padEnd(classWidth),
      ...statShades.map((shade) =>
        (row.shades.get(shade) ?? ".").toString().padStart(shadeWidth),
      ),
    ].join("  "),
  );
}
