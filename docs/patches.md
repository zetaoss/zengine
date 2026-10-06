# MediaWiki Patches

`mwz/patches/*.patch` contains diffs that modify MediaWiki core and extensions. The image `base` stage applies them to the MediaWiki directory with `patch -p1 --fuzz=0`, in filename order. A patch that does not apply causes the build to fail. Patches include only changed lines, rather than whole files, so upgrading MediaWiki or an extension does not silently overwrite upstream changes.

| Patch | Target | Description |
| --- | --- | --- |
| `ko-i18n.patch` | `languages/i18n/ko.json` | Changes Recent Changes markers (`잔글`/`새글`/`봇` to `m`/`N`/`b`) and `skin-action-viewsource` (`원본 보기` to `소스`). |
| `ready-tooltip-accesskeys.patch` | `resources/src/mediawiki.page.ready/ready.js` | Omits access-key hints from tooltips. |
| `skin-defaults-css-vars.patch` | `resources/src/mediawiki.less/mediawiki.skin.defaults.less` | Replaces color values with CSS variables (`var(--…)`). `svelte/src/shared/assets/appcolor-mw.css` defines the values for light and dark themes. |
| `SyntaxHighlight-attributes.patch` | `extensions/SyntaxHighlight_GeSHi/includes/SyntaxHighlight.php` | Allows attributes such as `run`, `fold`, and `notebook` on `<syntaxhighlight>` for code execution and related features. |
| `MsUpload-style.patch` | `extensions/MsUpload/resources/MsUpload.less` | Uses an SVG icon for the upload button, inverts it in dark mode, and uses skin color variables. |

## Updating a patch

Extract the original file from the MediaWiki image (`mediawiki:<version>-fpm`) or the installed extension. Put it under `a/<path>` and `b/<path>`, edit the copy under `b`, then create the diff:

```sh
diff -u --label a/<path> --label b/<path> a/<path> b/<path> > mwz/patches/<name>.patch
```

If a patch no longer applies after upgrading MediaWiki or an extension, recreate the same change against the new upstream file.
