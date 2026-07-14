import { mount, unmount } from 'svelte'

import getCurrentTitle from '$lib/utils/getCurrentTitle'
import getRLCONF from '$lib/utils/rlconf'
import httpy from '$shared/utils/httpy'

import AiEditPanel from './aiedit/AiEditPanel.svelte'
import EditHeader from './EditHeader.svelte'
import { replaceWikiEditorContent } from './wikiEditor'

interface EnabledArticleTplItem {
  content: string
  id: number
  title: string
}

interface EnabledArticleTplResp {
  key: string
  value?: EnabledArticleTplItem[]
}

interface StartEditBoxOptions {
  aiEditPanelMountElement: HTMLDivElement
  editAreaMountElement: HTMLDivElement
  rootElement: HTMLDivElement
  editHeaderMountElement: HTMLDivElement
}

export function startEditBox({ aiEditPanelMountElement, editAreaMountElement, rootElement, editHeaderMountElement }: StartEditBoxOptions) {
  let editHeaderInstance: object | null = null
  let editHeaderPromise: Promise<boolean> | null = null
  let aiEditPanelInstance: object | null = null
  let toggleAiEdit: ((visible: boolean) => void) | null = null
  let isDisposed = false
  const boilerplateContentByTitle: Record<string, string> = {}
  const editFormLayoutClasses = ['flex-1', 'min-w-0']

  function isEditAction() {
    const { wgAction } = getRLCONF()
    return wgAction === 'edit' || wgAction === 'submit'
  }

  function isCreateAction() {
    const { wgArticleId } = getRLCONF()
    return !wgArticleId
  }

  function findSaveButton() {
    const selectors = [
      '#wpSave',
      'button[name="wpSave"]',
      'input[name="wpSave"]',
      '#wpSaveWidget button',
      '.oo-ui-processDialog-actions-primary button',
    ]

    for (const selector of selectors) {
      const saveButton = document.querySelector<HTMLInputElement | HTMLButtonElement>(selector)
      if (saveButton) return saveButton
    }
    return null
  }

  function findEditForm() {
    return document.querySelector<HTMLElement>('#editform')
  }

  function findWikiPreview() {
    return document.querySelector<HTMLElement>('#wikiPreview')
  }

  function findWikiEditorTop() {
    return document.querySelector<HTMLElement>('.wikiEditor-ui-top')
  }

  function toggleWikiPreview(visible: boolean) {
    const wikiPreview = findWikiPreview()
    if (!wikiPreview) return

    if (!visible) {
      if (wikiPreview.dataset.zetaPrevDisplay === undefined) {
        wikiPreview.dataset.zetaPrevDisplay = wikiPreview.style.display
      }
      wikiPreview.style.display = 'none'
      return
    }

    const prevDisplay = wikiPreview.dataset.zetaPrevDisplay
    if (prevDisplay !== undefined) {
      if (prevDisplay) wikiPreview.style.display = prevDisplay
      else wikiPreview.style.removeProperty('display')
      delete wikiPreview.dataset.zetaPrevDisplay
    }
  }

  function placeRootBeforeEditForm() {
    const editForm = findEditForm()
    if (!rootElement.parentElement || !editForm?.parentElement) return false

    const targetNode = editForm.parentElement === editAreaMountElement ? editAreaMountElement : editForm
    if (!targetNode.parentElement) return false

    if (rootElement.parentElement !== targetNode.parentElement || rootElement.nextSibling !== targetNode) {
      targetNode.parentElement.insertBefore(rootElement, targetNode)
    }
    return true
  }

  function placeEditAreaBeforeEditForm() {
    const editForm = findEditForm()
    if (!editAreaMountElement.parentElement || !editForm?.parentElement) return false

    if (editForm.parentElement === editAreaMountElement) return true

    if (editAreaMountElement.parentElement !== editForm.parentElement || editAreaMountElement.nextSibling !== editForm) {
      editForm.parentElement.insertBefore(editAreaMountElement, editForm)
    }
    editAreaMountElement.appendChild(editForm)
    for (const className of editFormLayoutClasses) {
      editForm.classList.add(className)
    }
    return true
  }

  function placeEditHeaderBeforeWikiEditorTop() {
    const wikiEditorTop = findWikiEditorTop()
    if (!wikiEditorTop) return false

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
    for (const key of Object.keys(boilerplateContentByTitle)) {
      delete boilerplateContentByTitle[key]
    }
    for (const row of rows) {
      if (!row.title.startsWith('틀:새문서틀')) continue
      boilerplateContentByTitle[row.title] = row.content ?? ''
    }
    return rows.map((row) => row.title).filter((title) => title.startsWith('틀:새문서틀'))
  }

  function applyBoilerplate(title: string) {
    const content = boilerplateContentByTitle[title]
    if (!content) return
    replaceWikiEditorContent(content)
  }

  async function injectEditHeader() {
    if (isDisposed || !isEditAction()) return false
    if (editHeaderInstance) return true
    if (editHeaderPromise) return false
    if (!toggleAiEdit) return false

    editHeaderPromise = (async () => {
      const titles = await fetchEnabledBoilerplates()
      if (isDisposed || editHeaderInstance) return true
      if (!placeEditHeaderBeforeWikiEditorTop()) return false

      editHeaderInstance = mount(EditHeader, {
        target: editHeaderMountElement,
        props: {
          autoApplyOnSelect: isCreateAction(),
          onToggleAiEdit: toggleAiEdit,
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

  function applyAiEditVisible(visible: boolean) {
    aiEditPanelMountElement.classList.toggle('hidden', !visible)
    toggleWikiPreview(!visible)
  }

  function ensureAiEditPanel() {
    if (isDisposed || !isEditAction()) return null

    placeEditAreaBeforeEditForm()

    const editForm = findEditForm()
    if (!editForm) return null

    if (!aiEditPanelInstance) {
      const { wgArticleId } = getRLCONF()
      aiEditPanelInstance = mount(AiEditPanel, {
        target: aiEditPanelMountElement,
        props: {
          pageId: wgArticleId || undefined,
          requestType: isCreateAction() ? 'create' : 'edit',
          title: getCurrentTitle(),
        },
      })
    }

    return { editForm }
  }

  function injectAiEdit() {
    ;(window as typeof window & { __aiEditDebug?: Record<string, unknown> }).__aiEditDebug = {
      stage: 'inject:start',
      action: getRLCONF().wgAction,
      hasWpSave: !!document.querySelector('#wpSave'),
      url: window.location.href,
    }

    if (!isEditAction()) return false

    const saveButton = findSaveButton()
    if (!saveButton) {
      toggleAiEdit = null
      ;(window as typeof window & { __aiEditDebug?: Record<string, unknown> }).__aiEditDebug = {
        stage: 'inject:no-save-button',
        action: getRLCONF().wgAction,
        hasWpSave: false,
        url: window.location.href,
      }
      return false
    }

    toggleAiEdit = (visible: boolean) => {
      if (!ensureAiEditPanel()) return
      applyAiEditVisible(visible)
    }
    ;(window as typeof window & { __aiEditDebug?: Record<string, unknown> }).__aiEditDebug = {
      stage: 'inject:done',
      action: getRLCONF().wgAction,
      hasWpSave: true,
      buttonId: 'zeta-ai-edit-button',
      url: window.location.href,
    }
    return true
  }

  placeRootBeforeEditForm()
  placeEditAreaBeforeEditForm()
  placeEditHeaderBeforeWikiEditorTop()

  let attempts = 0
  const maxAttempts = 20
  let aiEditReady = false
  let editHeaderReady = false
  const timer = window.setInterval(() => {
    placeRootBeforeEditForm()
    placeEditHeaderBeforeWikiEditorTop()
    aiEditReady = injectAiEdit() || aiEditReady
    void injectEditHeader().then((mounted) => {
      editHeaderReady = mounted || editHeaderReady
    })
    attempts += 1
    if ((aiEditReady && editHeaderReady) || attempts >= maxAttempts) {
      window.clearInterval(timer)
    }
  }, 100)

  return () => {
    isDisposed = true
    window.clearInterval(timer)
    const editForm = findEditForm()
    if (editForm && editAreaMountElement.parentElement) {
      editAreaMountElement.parentElement.insertBefore(editForm, editAreaMountElement)
    }
    if (editForm) {
      for (const className of editFormLayoutClasses) {
        editForm.classList.remove(className)
      }
    }
    if (editHeaderMountElement.parentElement !== rootElement || rootElement.children[1] !== editHeaderMountElement) {
      rootElement.insertBefore(editHeaderMountElement, rootElement.children[1] ?? null)
    }
    toggleWikiPreview(true)
    if (editHeaderInstance) {
      unmount(editHeaderInstance)
      editHeaderInstance = null
    }
    editHeaderPromise = null
    if (aiEditPanelInstance) {
      unmount(aiEditPanelInstance)
      aiEditPanelInstance = null
    }
    toggleAiEdit = null
  }
}
