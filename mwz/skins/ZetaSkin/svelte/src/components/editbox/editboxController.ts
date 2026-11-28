import { mount, unmount } from 'svelte'

import getCurrentTitle from '$lib/utils/getCurrentTitle'
import mwapi from '$lib/utils/mwapi'
import getRLCONF from '$lib/utils/rlconf'
import { useRetrier } from '$shared/utils/retrier'
import httpy from '$shared/utils/httpy'

import EditHeader from './EditHeader.svelte'
import { replaceWikiEditorContent } from './wikiEditor'

interface EnabledArticleTplItem {
  id: number
  title: string
}

interface PreSaveTransformResp {
  parse?: {
    text?: string
  }
}

interface EnabledArticleTplResp {
  key: string
  value?: EnabledArticleTplItem[]
}

interface StartEditBoxOptions {
  rootElement: HTMLDivElement
  editHeaderMountElement: HTMLDivElement
}

export function startEditBox({ rootElement, editHeaderMountElement }: StartEditBoxOptions) {
  let editHeaderInstance: object | null = null
  let editHeaderPromise: Promise<boolean> | null = null
  let isDisposed = false
  const boilerplateContentByTitle: Record<string, string> = {}
  let boilerplateRequestId = 0

  function isEditAction() {
    const { wgAction } = getRLCONF()
    return wgAction === 'edit' || wgAction === 'submit'
  }

  function isCreateAction() {
    const { wgArticleId } = getRLCONF()
    return !wgArticleId
  }

  function findEditForm() {
    return document.querySelector<HTMLElement>('#editform')
  }

  function findWikiEditorTop() {
    return document.querySelector<HTMLElement>('.wikiEditor-ui-top')
  }

  function placeRootBeforeEditForm() {
    const editForm = findEditForm()
    if (!rootElement.parentElement || !editForm?.parentElement) return false

    if (rootElement.parentElement !== editForm.parentElement || rootElement.nextSibling !== editForm) {
      editForm.parentElement.insertBefore(rootElement, editForm)
    }
    return true
  }

  let hasLoggedMissingWikiEditor = false

  function placeEditHeader() {
    const wikiEditorTop = findWikiEditorTop()
    if (!wikiEditorTop) {
      if (!hasLoggedMissingWikiEditor) {
        console.debug('WikiEditor was not found yet. Retrying...')
        hasLoggedMissingWikiEditor = true
      }
      return false
    }

    hasLoggedMissingWikiEditor = false
    if (editHeaderMountElement.parentElement !== wikiEditorTop || wikiEditorTop.firstChild !== editHeaderMountElement) {
      wikiEditorTop.insertBefore(editHeaderMountElement, wikiEditorTop.firstChild)
    }
    return true
  }

  async function fetchEnabledBoilerplates() {
    const [data, err] = await httpy.get<EnabledArticleTplResp>('/api/article-tpl/enabled')
    if (err) {
      console.error(err)
      return []
    }
    const rows = Array.isArray(data?.value) ? data.value : []
    return rows.map((row) => row.title).filter((title) => title.startsWith('틀:새문서틀'))
  }

  async function applyBoilerplate(title: string) {
    const requestId = ++boilerplateRequestId
    if (!title) return

    let content = boilerplateContentByTitle[title]
    if (content === undefined) {
      const [data, err] = await mwapi.get<PreSaveTransformResp>({
        action: 'parse',
        onlypst: true,
        text: `{{subst:${title}}}`,
        title: getCurrentTitle(),
      })
      if (err) {
        console.error('Failed to substitute boilerplate template:', err)
        return
      }
      content = data?.parse?.text ?? ''
      boilerplateContentByTitle[title] = content
    }

    if (requestId !== boilerplateRequestId) return
    replaceWikiEditorContent(content)
  }

  async function injectEditHeader() {
    if (isDisposed || !isEditAction()) return false
    if (editHeaderInstance) return true
    if (editHeaderPromise) return false

    editHeaderPromise = (async () => {
      const titles = await fetchEnabledBoilerplates()
      if (isDisposed || editHeaderInstance) return true
      if (!placeEditHeader()) return false

      editHeaderInstance = mount(EditHeader, {
        target: editHeaderMountElement,
        props: {
          autoApplyOnSelect: isCreateAction(),
          onSelect: applyBoilerplate,
          titles,
        },
      })
      return true
    })()

    try {
      return await editHeaderPromise
    } finally {
      editHeaderPromise = null
    }
  }

  // Only edit pages have the edit form and WikiEditor; elsewhere the retries below would only search for them.
  if (!isEditAction()) return () => {}

  const retrier = useRetrier(
    () => {
      if (isDisposed) return
      placeRootBeforeEditForm()
      if (!placeEditHeader()) {
        retrier.schedule()
        return
      }
      void injectEditHeader()
        .then((mounted) => {
          if (mounted) retrier.clear()
          else retrier.schedule()
        })
        .catch((err) => {
          console.error('Failed to inject edit header:', err)
          retrier.schedule()
        })
    },
    100,
    5000,
    1.5,
    20,
  )

  retrier.start()

  return () => {
    isDisposed = true
    retrier.clear()
    if (editHeaderMountElement.parentElement !== rootElement) {
      rootElement.appendChild(editHeaderMountElement)
    }
    if (editHeaderInstance) {
      unmount(editHeaderInstance)
      editHeaderInstance = null
    }
    editHeaderPromise = null
  }
}
