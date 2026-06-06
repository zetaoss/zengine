import { mount, unmount } from 'svelte'

import getCurrentTitle from '$lib/utils/getCurrentTitle'
import getRLCONF from '$lib/utils/rlconf'
import httpy from '$shared/utils/httpy'

import AiEditPanel from './aiedit/AiEditPanel.svelte'
import TemplateBox from './TemplateBox.svelte'

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
  templateBoxMountElement: HTMLDivElement
}

export function startEditBox({
  aiEditPanelMountElement,
  editAreaMountElement,
  rootElement,
  templateBoxMountElement,
}: StartEditBoxOptions) {
  let templateSelectInstance: object | null = null
  let templateSelectPromise: Promise<boolean> | null = null
  let aiEditPanelInstance: object | null = null
  let aiEditVisible = false
  let toggleAiEdit: (() => void) | null = null
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

  function findSaveWidget(saveButton: HTMLInputElement | HTMLButtonElement) {
    const saveWidget = document.getElementById('wpSaveWidget')
    if (saveWidget) return saveWidget
    return saveButton.closest<HTMLElement>('.oo-ui-buttonInputWidget')
  }

  function findEditForm() {
    return document.querySelector<HTMLElement>('#editform')
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
    const textarea = document.querySelector<HTMLTextAreaElement>('#wpTextbox1')
    if (!textarea) return

    if (title === '') {
      textarea.focus()
      textarea.value = ''
      textarea.dispatchEvent(new Event('input', { bubbles: true }))
      return
    }

    const content = boilerplateContentByTitle[title] ?? ''
    if (!content) return

    textarea.focus()
    textarea.setSelectionRange(0, textarea.value.length)
    const inserted = document.execCommand('insertText', false, content)

    if (!inserted) {
      textarea.value = content
      textarea.dispatchEvent(new Event('input', { bubbles: true }))
    }
  }

  async function injectBoilerplateBox() {
    if (isDisposed || !isEditAction()) return false
    if (templateSelectInstance) return true
    if (templateSelectPromise) return false
    if (!toggleAiEdit) return false

    templateSelectPromise = (async () => {
      const titles = await fetchEnabledBoilerplates()
      if (isDisposed || templateSelectInstance) return true

      templateSelectInstance = mount(TemplateBox, {
        target: templateBoxMountElement,
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
      return await templateSelectPromise
    } finally {
      templateSelectPromise = null
    }
  }

  function applyAiEditVisible(
    editForm: HTMLElement | null,
    saveButton: HTMLInputElement | HTMLButtonElement,
    visible: boolean,
    hideEditForm: boolean,
  ) {
    aiEditVisible = visible
    aiEditPanelMountElement.classList.toggle('hidden', !visible)
    editForm?.classList.toggle('hidden', visible && hideEditForm)
    const hideSaveButton = visible && hideEditForm
    const saveWidget = findSaveWidget(saveButton)
    if (saveWidget) {
      saveWidget.classList.toggle('hidden', hideSaveButton)
      return
    }
    saveButton.classList.toggle('hidden', hideSaveButton)
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

    toggleAiEdit = () => {
      const aiEditPanel = ensureAiEditPanel()
      if (!aiEditPanel) return
      if (aiEditVisible) {
        applyAiEditVisible(aiEditPanel.editForm, saveButton, false, isCreateAction())
        return
      }
      applyAiEditVisible(aiEditPanel.editForm, saveButton, true, isCreateAction())
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

  let attempts = 0
  const maxAttempts = 20
  let aiEditReady = false
  let boilerplateReady = false
  const timer = window.setInterval(() => {
    placeRootBeforeEditForm()
    aiEditReady = injectAiEdit() || aiEditReady
    void injectBoilerplateBox().then((mounted) => {
      boilerplateReady = mounted || boilerplateReady
    })
    attempts += 1
    if ((aiEditReady && boilerplateReady) || attempts >= maxAttempts) {
      window.clearInterval(timer)
    }
  }, 100)

  return () => {
    isDisposed = true
    window.clearInterval(timer)
    const saveButton = findSaveButton()
    if (saveButton) {
      const saveWidget = findSaveWidget(saveButton)
      if (saveWidget) saveWidget.classList.remove('hidden')
      else saveButton.classList.remove('hidden')
    }
    const editForm = findEditForm()
    if (editForm && editAreaMountElement.parentElement) {
      editAreaMountElement.parentElement.insertBefore(editForm, editAreaMountElement)
    }
    if (editForm) {
      for (const className of editFormLayoutClasses) {
        editForm.classList.remove(className)
      }
    }
    editForm?.classList.remove('hidden')
    if (templateSelectInstance) {
      unmount(templateSelectInstance)
      templateSelectInstance = null
    }
    templateSelectPromise = null
    if (aiEditPanelInstance) {
      unmount(aiEditPanelInstance)
      aiEditPanelInstance = null
    }
    toggleAiEdit = null
  }
}
