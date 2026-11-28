# Agent Working Guide (/app/mwz/skins/ZetaSkin)

## Scope

- This guide applies to the ZetaSkin MediaWiki skin and its embedded Svelte app.
- For repo-wide guidance, read `/app/AGENTS.md` first.

## Skin Svelte Dev Runtime

- In local dev, `dev2` should run the Vite dev server path for `mwz/skins/ZetaSkin/svelte`.
- The live runtime source is `mwz/skins/ZetaSkin/svelte/src/main.ts` and its imported modules.
- `mwz/skins/ZetaSkin/dist/app.js` and `mwz/skins/ZetaSkin/dist/app.css` are compatibility stubs for MediaWiki asset injection, not the authoritative dev source.
- A production build (`vite build`) writes `dist/app.<hash>.js`, `dist/app.<hash>.css` and `dist/assets.php`, which `SkinZetaSkin` requires for the asset URLs. The dev server writes its own `assets.php` pointing at the stubs.
- If browser behavior looks stale, check Vite/HMR and browser cache behavior before assuming `dist/` is out of date.

## Menus

- `SkinZetaSkin::onSkinTemplateNavigation__Universal` copies the page menus into the `menu` JS variable, which the Svelte app renders. The subject and talk links come from `associated-pages` (in MediaWiki 1.43 and later; 1.47 no longer has `namespaces`), stored under the `namespaces` key that `getPageButtons` reads.
- `skin.json` declares the menus the skin renders (`menus`: `associated-pages`, `views`, `actions`, `user-menu`). Without it MediaWiki 1.47 warns that skins must pass `menus` and builds the deprecated `personal` menu. The template itself renders only the category portlets.

## Shared Code

- `mwz/skins/ZetaSkin/svelte/src/shared/` is a symlink to `svelte/src/shared/`.
- Edit shared utilities, components, and styles in `svelte/src/shared/`, not through the symlink under `mwz/`.
- Only code the main app also imports belongs there. Skin-only code goes in `mwz/skins/ZetaSkin/svelte/src/lib/` (`hack/check-shared.mjs` enforces this).

## Verification

- For skin Svelte changes, run `pnpm type-check` and the relevant `eslint` commands before finalizing. Unit tests (`src/**/*.test.ts`) run with `pnpm test:unit`.
- Treat `svelte-check` warnings as issues to resolve or explicitly document.

