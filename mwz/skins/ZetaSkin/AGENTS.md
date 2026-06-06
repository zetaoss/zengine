# Agent Working Guide (/app/mwz/skins/ZetaSkin)

## Scope

- This guide applies to the ZetaSkin MediaWiki skin and its embedded Svelte app.
- For repo-wide guidance, read `/app/AGENTS.md` first.

## Skin Svelte Dev Runtime

- In local dev, `dev2` should run the Vite dev server path for `mwz/skins/ZetaSkin/svelte`.
- The live runtime source is `mwz/skins/ZetaSkin/svelte/src/main.ts` and its imported modules.
- `mwz/skins/ZetaSkin/dist/app.js` and `mwz/skins/ZetaSkin/dist/app.css` are compatibility stubs for MediaWiki asset injection, not the authoritative dev source.
- If browser behavior looks stale, check Vite/HMR and browser cache behavior before assuming `dist/` is out of date.

## Shared Code

- `mwz/skins/ZetaSkin/svelte/src/shared/` is a symlink to `svelte/src/shared/`.
- Edit shared utilities, components, and styles in `svelte/src/shared/`, not through the symlink under `mwz/`.

## Verification

- For skin Svelte changes, run `pnpm type-check` and the relevant `eslint` commands before finalizing.
- Treat `svelte-check` warnings as issues to resolve or explicitly document.

