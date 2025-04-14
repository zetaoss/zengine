import { resolve } from "path";
import { readFileSync } from "fs";

const file1 = resolve('mwz/skins/ZetaSkin/vue/package.json');
const file2 = resolve('vue/package.json');

const json1 = JSON.parse(readFileSync(file1, 'utf-8'));
const json2 = JSON.parse(readFileSync(file2, 'utf-8'));

const deps1 = { ...json1.dependencies, ...json1.devDependencies };
const deps2 = { ...json2.dependencies, ...json2.devDependencies };

const allDeps = Array.from(new Set([...Object.keys(deps1), ...Object.keys(deps2)]));

const pad = (str, len) => (str || "").padEnd(len);
const SEP = '   ';
const gray = str => `\x1b[90m${str}\x1b[0m`;
const red = str => `\x1b[91m${str}\x1b[0m`;
const green = str => `\x1b[92m${str}\x1b[0m`;

const compareVersions = (v1, v2) => {
    if (!v1 || !v2) return ' ';
    if (v1 === v2) return '=';
    return v1 > v2 ? '>' : '<';
};

const rows = allDeps.map(dep => {
    const v1 = deps1[dep] || "";
    const v2 = deps2[dep] || "";
    return { dep, v1, v2, cmp: compareVersions(v1, v2) };
});

const cmpOrder = { '<': 0, '>': 1, '=': 2, ' ': 3 };
rows.sort((a, b) => cmpOrder[a.cmp] - cmpOrder[b.cmp] || a.dep.localeCompare(b.dep));

const maxDepLen = Math.max(...rows.map(r => r.dep.length), "DEP".length);
const maxV1Len = Math.max(...rows.map(r => r.v1.length), "FILE1".length);
const maxV2Len = Math.max(...rows.map(r => r.v2.length), "FILE2".length);

console.log(`FILE1: ${file1}`);
console.log(`FILE2: ${file2}\n`);
console.log(`${pad("DEP", maxDepLen)}${SEP}${pad("FILE1", maxV1Len)}${SEP}CMP${SEP}${pad("FILE2", maxV2Len)}`);

for (const { dep, v1, v2, cmp } of rows) {
    const line = `${pad(dep, maxDepLen)}${SEP}${pad(v1, maxV1Len)}${SEP} ${cmp} ${SEP}${pad(v2, maxV2Len)}`;
    console.log(!v1 || !v2 ? gray(line) : cmp === '=' ? green(line) : red(line));
}
