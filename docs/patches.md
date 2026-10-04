# MediaWiki 패치

`mwz/patches/*.patch`는 MediaWiki 본체와 확장을 고치는 diff다. 이미지 `base` 단계가 MediaWiki 디렉터리에 `patch -p1 --fuzz=0`으로 적용한다(파일 이름 순). 맞지 않는 패치가 있으면 빌드가 실패한다. MediaWiki나 확장을 올릴 때 조용히 예전 파일로 덮어쓰지 않도록 파일 전체가 아니라 바꾼 부분만 둔다.

| 패치 | 대상 | 내용 |
| --- | --- | --- |
| `ko-i18n.patch` | `languages/i18n/ko.json` | 최근 바뀜의 표시 글자(`잔글`/`새글`/`봇` → `m`/`N`/`b`), `skin-action-viewsource`(`원본 보기` → `소스`) |
| `ready-tooltip-accesskeys.patch` | `resources/src/mediawiki.page.ready/ready.js` | 툴팁에 접근 키 안내를 붙이지 않는다 |
| `skin-defaults-css-vars.patch` | `resources/src/mediawiki.less/mediawiki.skin.defaults.less` | 색상 변수를 CSS 변수(`var(--…)`)로 바꾼다. 값은 `svelte/src/shared/assets/appcolor-mw.css`가 라이트/다크별로 정의한다 |
| `SyntaxHighlight-attributes.patch` | `extensions/SyntaxHighlight_GeSHi/includes/SyntaxHighlight.php` | `<syntaxhighlight>`에 코드 실행 등에 쓰는 속성(`run`, `fold`, `notebook` 등)을 허용한다 |
| `MsUpload-style.patch` | `extensions/MsUpload/resources/MsUpload.less` | 업로드 버튼을 SVG 아이콘으로, 다크 모드 반전, 색상을 skin 변수로 |

## 패치 고치기

MediaWiki 이미지(`mediawiki:<버전>-fpm`)나 설치된 확장에서 원본을 꺼내 `a/<경로>`와 `b/<경로>`에 두고 `b`를 고친 뒤 diff를 만든다.

```sh
diff -u --label a/<경로> --label b/<경로> a/<경로> b/<경로> > mwz/patches/<이름>.patch
```

MediaWiki나 확장을 올렸는데 패치가 맞지 않으면, 새 원본에서 같은 변경을 다시 만든다.
