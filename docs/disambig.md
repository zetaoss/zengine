# 동음이의 기능 가이드

동음이의 기능은 일반 이름공간의 문서 중 같은 기본 제목을 공유하는 문서를 `Disambig` 이름공간의 문서 하나로 묶고, 각 일반 문서 상단에 다른 뜻 목록을 표시한다. 토론 문서를 비롯한 다른 이름공간은 동음이의 대상으로 처리하지 않는다. 동음이의 문서가 원본 데이터이며, ZetaExtension이 표시용 캐시와 문서 관계 인덱스를 관리하고 ZetaSkin이 이를 렌더링한다.

## 문서 모델

ZetaExtension은 다음 이름공간을 등록한다.

| 이름공간 | ID | 용도 |
| --- | ---: | --- |
| `Disambig` | 3002 | 동음이의 문서 |
| `Disambig_talk` | 3003 | 동음이의 토론 문서 |

일반 문서 제목 끝에 괄호가 있으면 마지막 괄호 부분을 제외한 값을 기본 제목으로 사용한다.

```text
사과       -> Disambig:사과
사과 (과일) -> Disambig:사과
사과(기업)  -> Disambig:사과
```

동음이의 문서는 첫 번째 글머리표 목록의 각 항목을 다음 형식으로 표현한다.

```wikitext
* [[사과 (과일)|사과 (과일)]] - 과일
* [[사과 (기업)]] - 기업
```

각 목록 항목에서 직접 연결된 첫 번째 링크가 대상 문서가 된다. 링크 문구는 화면에 표시할 이름이고 ` - ` 뒤의 문자열은 설명이다. UI 편집기는 위 형식으로만 문서를 읽고 저장하므로 동음이의 문서에 다른 위키텍스트가 섞여 있으면 표 형식 편집을 거부한다.

## 사용자 흐름

일반 이름공간의 존재하는 문서를 조회하면 ZetaExtension이 기본 제목과 대응하는 `Disambig` 문서 정보를 `disambigRegistration`으로 전달한다.

- 관계가 아직 없으면 페이지 메뉴에서 동음이의 생성 또는 등록 모달을 열 수 있다.
- 새 동음이의 문서를 만들 때 같은 기본 제목을 가진 일반 문서와 기존 `{{다른 뜻}}` 틀의 대상을 자동으로 탐색한다.
- 기존 동음이의 문서는 현재 내용을 표로 불러와 항목 추가, 삭제, 정렬 및 설명 수정을 지원한다.
- 저장은 MediaWiki edit API를 사용한다. 기존 문서 편집에는 `basetimestamp`를 전달해 충돌을 감지한다.
- 동음이의 관계가 있는 일반 문서에는 문서 상단에 목록과 편집 버튼을 표시한다.
- 모달에서 동음이의 문서를 삭제하면 연결 관계도 페이지 삭제 훅을 통해 제거된다.

항목 제목은 기본 제목 자체이거나 같은 기본 제목 뒤에 괄호가 붙은 형태여야 한다. 예를 들어 기본 제목이 `사과`이면 `사과`, `사과 (과일)`, `사과(기업)`은 허용되지만 `풋사과`는 허용되지 않는다.

## 구성 요소

| 역할 | 경로 |
| --- | --- |
| 이름공간, 훅, Job 등록 | `mwz/extensions/ZetaExtension/extension.json` |
| 기본 제목 및 등록 정보 계산 | `mwz/extensions/ZetaExtension/includes/Disambig/DisambigHooks.php` |
| 파싱, 캐시 및 관계 동기화 | `mwz/extensions/ZetaExtension/includes/Disambig/DisambigService.php` |
| 저장·삭제·복구·이동 훅 연결 | `mwz/extensions/ZetaExtension/includes/Collection/CollectionHooks.php` |
| 레거시 틀과 대표 제목 정리 | `mwz/extensions/ZetaExtension/includes/Disambig/CleanupLegacyDisambigJob.php` |
| 전체 캐시 재구축 | `mwz/extensions/ZetaExtension/maintenance/RebuildDisambigs.php` |
| 페이지 데이터 조회 | `mwz/skins/ZetaSkin/includes/PageDataProvider.php` |
| 상단 목록 | `mwz/skins/ZetaSkin/svelte/src/components/disambig/DisambigApex.svelte` |
| 생성·편집 모달 | `mwz/skins/ZetaSkin/svelte/src/components/disambig/DisambigModal.svelte` |

## 저장 구조

현재 테이블은 `ldb` 데이터베이스에 있다.

### `disambigs`

동음이의 문서별 표시용 캐시를 저장한다.

| 열 | 의미 |
| --- | --- |
| `id` | `Disambig` 문서의 MediaWiki page ID |
| `cache` | `id`, `text`, `nodes`를 담은 JSON |
| `entries` | 캐시의 항목 수 |
| `created_at` | 최초 생성 시각 |
| `updated_at` | 마지막 동기화 시각 |

캐시의 각 node에는 제목, 링크 문구, URL, 설명과 함께 존재하는 문서의 `id` 또는 아직 없는 문서를 나타내는 `new`가 들어간다.

### `disambig_pages`

일반 문서 제목과 동음이의 문서의 관계를 저장한다.

| 열 | 의미 |
| --- | --- |
| `disambig_id` | 관계를 소유한 `Disambig` 문서 ID |
| `page_title` | 일반 이름공간 대상 문서의 정규화된 DB 제목 |
| `page_id` | 대상 문서가 존재할 때의 page ID, 미생성 문서는 `NULL` |

제목 관계를 미리 저장하므로 동음이의 항목이 아직 존재하지 않아도 나중에 일반 이름공간에 같은 제목의 문서가 생성되면 관계를 다시 계산할 수 있다. 현재 스키마는 `page_title`을 기본 키로 사용하고 `page_id`도 유일하게 유지하므로 한 일반 문서는 하나의 동음이의 문서에만 연결된다.

두 테이블의 생성 SQL과 기존 개발 스키마를 일반 이름공간 전용 구조로 변환하는 호환 처리는 `goapp/app/database/migrations/202608010001_collection_page_indexes.up.sql`에서 함께 관리한다. 테이블을 MediaWiki 기본 DB로 이전하거나 Extension 스키마 업데이트로 관리하도록 변경할 경우, 서비스와 Skin의 `ldb.` 참조 및 배포 시 마이그레이션 실행 주체를 함께 변경해야 한다.

## 일회성 `202608010001_collection_page_indexes` 마이그 후속 작업

새 application image에 포함된 `tool`로 마이그레이션을 적용한다.

```bash
tool migrate
```

처음 적용한 뒤에는 기존 Binder 문서를 모두 다시 파싱해 과거 redlink와 redirect dependency를 채운다.

```bash
MW_INSTALL_PATH=/app/w php /app/mwz/extensions/ZetaExtension/maintenance/RebuildBinders.php --server localhost
```

구버전 PHP와 변경된 `binder_pages` 스키마는 서로 호환되지 않으므로 마이그레이션과 소스 전환 중에는 MediaWiki 쓰기 요청과 관련 Job 처리를 중단한다.

## 동기화 과정

`CollectionHooks`가 MediaWiki 페이지 생명주기를 `DisambigService`에 연결한다.

```text
Disambig 문서 저장
  -> 첫 번째 목록 파싱
  -> disambig_pages 관계 교체
  -> disambigs 캐시 upsert
  -> 레거시 정리 Job enqueue

일반 문서 저장·복구
  -> 동일한 미생성 제목 관계 탐색
  -> 관련 Disambig 캐시 재구축

문서 이동
  -> 이전 page ID 관계 재계산
  -> 새 제목 및 생성된 redirect 관계 재계산

문서 삭제
  -> Disambig 문서이면 캐시와 모든 관계 삭제
  -> 일반 문서이면 관련 캐시 재구축
```

관계와 캐시는 하나의 atomic section에서 갱신된다. 캐시는 파생 데이터이므로 내용이 의심스러울 때 동음이의 문서를 다시 저장하거나 전체 재구축 명령을 실행할 수 있다.

## 레거시 동음이의 정리

동음이의 문서를 저장하면 `zetaCleanupLegacyDisambig` Job이 연결된 일반 문서별로 등록된다.

- 일반 문서 본문의 줄 단위 `{{다른 뜻}}` 또는 `{{다른_뜻}}` 틀을 제거한다.
- 기본 제목 문서가 목록에 없고, 해당 문서가 없거나 redirect이면 첫 번째 동음이의 항목으로 redirect를 생성하거나 갱신한다.
- 이미 일반 본문이 있는 기본 제목 문서는 덮어쓰지 않는다.
- 편집은 동음이의 문서를 저장한 사용자의 권한으로 수행하고 minor edit로 기록한다.
- 같은 대상의 중복 Job은 Job queue 옵션으로 제거한다.

Job 실행 상태와 실패는 MediaWiki Job queue에서 확인한다. 편집 사용자를 복원할 수 없거나 revision 저장이 실패하면 Job은 실패를 반환한다.

## 전체 재구축

모든 `Disambig` 문서를 다시 파싱해 관계와 캐시를 재구축한다.

```bash
MW_INSTALL_PATH=/app/w php /app/mwz/extensions/ZetaExtension/maintenance/RebuildDisambigs.php
```

명령은 성공한 문서를 한 줄씩 출력하고 마지막에 성공 개수를 표시한다. 하나라도 실패하면 실패한 page ID, 제목과 오류를 모아 non-zero 상태로 종료한다.

다음 상황에서 전체 재구축을 고려한다.

- 테이블을 새로 생성하거나 복구한 뒤
- 파싱 또는 캐시 형식을 변경한 뒤
- 훅이 비활성화된 동안 동음이의 문서가 변경된 뒤
- `disambig_pages.page_id`와 실제 MediaWiki page ID가 일치하지 않는다고 의심될 때

## 점검 순서

화면에 동음이의 목록이 나타나지 않으면 다음 순서로 확인한다.

1. 대응하는 `Disambig:<기본 제목>` 문서가 존재하고 지원되는 글머리표 형식인지 확인한다.
2. `disambigs`에 해당 Disambig page ID의 유효한 JSON 캐시가 있는지 확인한다.
3. `disambig_pages`에 현재 문서의 `page_id` 또는 정규화된 제목 관계가 있는지 확인한다.
4. `PageSaveComplete`, `PageDeleteComplete`, `PageUndeleteComplete`, `PageMoveComplete` 훅과 `zetaCleanupLegacyDisambig` Job 등록 상태를 확인한다.
5. 필요한 경우 전체 재구축 명령을 실행하고 실패 문서를 확인한다.
