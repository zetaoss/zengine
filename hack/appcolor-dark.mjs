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
const DARK_OUTPUT_PATH = path.join(OUTPUT_DIR, "appcolor-dark.css");

async function main() {
  try {
    const content = await readFile(THEME_CSS_PATH, "utf8");
    const lines = content.split("\n");

    const families = {};
    const colorData = [];

    const shadeMap = {
      "50": "950",
      "100": "900",
      "200": "800",
      "300": "700",
      "400": "600",
      "500": "500",
      "600": "400",
      "700": "300",
      "800": "200",
      "900": "100",
      "950": "50",
    };

    const oklchFullRegex =
      /^\s*(--color-([a-z0-9-]+)-([\d]+)):\s*oklch\(([\d.]+)%\s+([\d.]+)\s+([\d.]+)\);/;

    for (const line of lines) {
      const matchFull = line.match(oklchFullRegex);
      if (matchFull) {
        const [full, variable, family, shade, lStr, c, h] = matchFull;
        if (!families[family]) families[family] = {};
        families[family][shade] = { lStr, c, h };
        colorData.push({ variable, family, shade });
        continue;
      }
    }

    // Generate Dark Mode Content
    const darkLines = [];
    darkLines.push("  --color-white: #111;");
    darkLines.push("  --color-black: #fff;");

    for (const { variable, family, shade } of colorData) {
      const targetShade = shadeMap[shade];
      if (targetShade && families[family][targetShade]) {
        const { lStr, c, h } = families[family][targetShade];
        darkLines.push(`  ${variable}: oklch(${lStr}% ${c} ${h});`);
      }
    }

    const darkOutputContent = `.dark {\n${darkLines.join("\n")}\n}\n`;
    await mkdir(OUTPUT_DIR, { recursive: true });
    await writeFile(DARK_OUTPUT_PATH, darkOutputContent, "utf8");
    console.log(
      `Successfully generated ${DARK_OUTPUT_PATH} using full LCH shade-to-shade mapping`,
    );
  } catch (err) {
    console.error(`Error:`, err.message);
    process.exit(1);
  }
}

main();
