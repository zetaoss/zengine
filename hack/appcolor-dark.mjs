#!/usr/bin/env node

import { readFile, writeFile, mkdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const THEME_CSS_PATH = path.resolve(
  SCRIPT_DIR,
  "../svelte/node_modules/tailwindcss/theme.css",
);
const OUTPUT_DIR = path.resolve(SCRIPT_DIR, "../svelte/src/shared/assets");
const OUTPUT_PATH = path.join(OUTPUT_DIR, "appcolor-dark.css");

async function main() {
  try {
    const content = await readFile(THEME_CSS_PATH, "utf8");
    const lines = content.split("\n");

    const oklchRegex =
      /(--color-[a-z0-9-]+):\s*oklch\(([\d.]+)%\s+([\d.]+)\s+([\d.]+)\);/;

    const darkLines = ["  --color-white: #111;", "  --color-black: #fff;"];

    const L_MIN = 15;

    for (const line of lines) {
      const match = line.match(oklchRegex);
      if (match) {
        const [full, variable, lStr, c, h] = match;
        const L = parseFloat(lStr);
        // Map L [0, 100] to inverted range [L_MIN, 100]
        // L=0 -> 100, L=100 -> L_MIN
        const newL = (L_MIN + ((100 - L) * (100 - L_MIN)) / 100).toFixed(1);
        darkLines.push(`  ${variable}: oklch(${newL}% ${c} ${h});`);
      }
    }

    const outputContent = `.dark {\n${darkLines.join("\n")}\n}\n`;

    await mkdir(OUTPUT_DIR, { recursive: true });
    await writeFile(OUTPUT_PATH, outputContent, "utf8");

    console.log(`Successfully generated ${OUTPUT_PATH} (Lightness Inversion)`);
  } catch (err) {
    console.error(`Error:`, err.message);
    process.exit(1);
  }
}

main();
