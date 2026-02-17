<svelte:options runes={true} />

<script lang="ts">
  import { onMount, tick } from 'svelte'

  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import useAuthStore from '$lib/stores/auth'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZCard from '$shared/ui/ZCard.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'
  import { sha256Hex } from '$shared/utils/sha256'

  import { maskEmail } from './util/mask'

  type AvatarType = 1 | 2 | 3

  const auth = useAuthStore()
  const { isLoggedIn, userInfo } = auth

  let userName = $derived(page.params.user_name ?? '')
  let encodedUsername = $derived(userName.replace(/ /g, '_'))

  let isLoading = $state(false)
  let loadError = $state<string | null>(null)

  let saving = $state(false)
  let saveError = $state<string | null>(null)
  let saveOk = $state(false)

  let isMe = $derived($isLoggedIn && ($userInfo?.name ?? '') === userName)

  let selectedType = $state<AvatarType>(1)

  let gravatarEmail = $state('')
  let gravatarHint = $state('')
  let initialGravatarHint = $state('')
  let verifiedGravatarHash = $state('')
  let verifiedGravatarHint = $state('')
  let gravatarError = $state<string | null>(null)

  let isEditingGravatar = $state(false)
  let gravatarInput = $state<HTMLInputElement | null>(null)
  let gravatarBusy = $state(false)

  let canEditGravatar = $derived(selectedType === 3)
  let canSave = $derived(isMe && !saving)

  function validateEmail(email: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())
  }

  let gravatarDisplay = $derived(isEditingGravatar ? gravatarEmail : gravatarHint)
  let gravatarChanged = $derived(gravatarEmail.trim() !== '')

  async function startEditGravatar() {
    if (!canEditGravatar) return

    saveOk = false
    saveError = null
    gravatarError = null
    isEditingGravatar = true
    gravatarEmail = ''

    await tick()
    gravatarInput?.focus()
  }

  function cancelEditGravatar() {
    gravatarEmail = ''
    isEditingGravatar = false
    saveOk = false
    saveError = null
    gravatarError = null
  }

  async function confirmGravatar() {
    if (!canEditGravatar) return

    saveOk = false
    saveError = null
    gravatarError = null

    const email = gravatarEmail.trim()
    if (!validateEmail(email)) {
      gravatarError = 'Gravatar 이메일 형식이 올바르지 않습니다.'
      return
    }

    const [ghash, err] = await sha256Hex(email.toLowerCase())
    if (err) {
      gravatarError = 'Gravatar 해시 생성 실패'
      return
    }
    const ghint = maskEmail(email)

    gravatarBusy = true
    try {
      const [vdata, verr] = await httpy.get<{ ok: boolean; ghash?: string }>('/api/me/gravatar/verify', {
        ghash,
      })

      if (verr || !vdata?.ok) {
        gravatarError = 'Gravatar 확인 실패'
        return
      }

      verifiedGravatarHash = ghash
      verifiedGravatarHint = ghint
      gravatarHint = ghint
      initialGravatarHint = ghint
      gravatarEmail = ''
      isEditingGravatar = false
    } finally {
      gravatarBusy = false
    }
  }

  async function loadAvatarConfig() {
    const [cfg, cfgErr] = await httpy.get<{ t?: unknown; ghint?: unknown }>('/api/me/avatar')

    if (cfgErr) {
      selectedType = 1
      gravatarHint = ''
      initialGravatarHint = ''
      verifiedGravatarHash = ''
      verifiedGravatarHint = ''
      return
    }

    const rawType = Number(cfg?.t ?? 1)
    selectedType = rawType === 2 || rawType === 3 ? rawType : 1

    const rawHint = cfg?.ghint ?? ''
    gravatarHint = String(rawHint)
    initialGravatarHint = gravatarHint
    verifiedGravatarHash = ''
    verifiedGravatarHint = ''
  }

  async function load() {
    loadError = null
    saveError = null
    saveOk = false
    gravatarError = null

    await auth.update()
    if (!$isLoggedIn) {
      loadError = 'failed to load profile'
      return
    }

    await loadAvatarConfig()
    isEditingGravatar = false
  }

  async function save() {
    if (!canSave) return

    saveError = null
    saveOk = false

    saving = true
    try {
      const payload: { t: AvatarType; ghash?: string; ghint?: string } = { t: selectedType }

      if (selectedType === 3 && verifiedGravatarHash !== '') {
        payload.ghash = verifiedGravatarHash
        payload.ghint = verifiedGravatarHint
      }

      const [, err] = await httpy.post('/api/me/avatar', payload)
      if (err) {
        saveError = '저장 실패'
        return
      }

      if ($userInfo?.id) {
        localStorage.setItem('v', String(Date.now()))
      }

      verifiedGravatarHash = ''
      verifiedGravatarHint = ''
      saveOk = true
      await auth.update()
      await loadAvatarConfig()
    } finally {
      saving = false
    }
  }

  let mounted = $state(false)
  let lastLoadedUserName = ''

  $effect(() => {
    if (selectedType) {
      saveOk = false
      saveError = null
      isEditingGravatar = false
      gravatarError = null
    }
  })

  $effect(() => {
    if (gravatarHint !== initialGravatarHint) {
      saveOk = false
    }
  })

  $effect(() => {
    if (!mounted) return

    const currentUserName = userName.trim()
    if (currentUserName === '' || currentUserName === lastLoadedUserName) return

    lastLoadedUserName = currentUserName
    isLoading = true

    void load().finally(() => {
      if (mounted && lastLoadedUserName === currentUserName) {
        isLoading = false
      }
    })
  })

  onMount(() => {
    mounted = true
    return () => {
      mounted = false
    }
  })
</script>

<div class="mx-auto max-w-4xl py-6">
  {#if isLoading}
    <div class="text-center">
      <ZSpinner />
    </div>
  {:else if loadError}
    <div class="mt-4 text-center text-sm text-red-500">
      {loadError}
    </div>
  {:else}
    <ZCard class="mt-4 p-6">
      <svelte:fragment slot="header">
        <div class="flex items-baseline justify-between">
          <span class="text-base font-semibold">프로필 편집</span>
          <a href={resolve(`/user/${encodedUsername}`)} class="text-sm">돌아가기</a>
        </div>
      </svelte:fragment>

      {#if !isMe}
        <div class="text-sm text-red-500">본인만 프로필을 수정할 수 있습니다.</div>
      {:else}
        <div class="space-y-4">
          <div class="text-sm font-semibold">아바타 선택</div>

          <ul class="space-y-2">
            <li class="z-ring flex items-center gap-3 rounded p-3 ring-1">
              <input
                type="radio"
                name="avatarType"
                checked={selectedType === 1}
                onchange={() => {
                  selectedType = 1
                }}
                class="accent-current"
              />
              <div class="flex items-center gap-3">
                {#if $userInfo}
                  <AvatarIcon user={$userInfo} size={60} typ={1} />
                {/if}
                <div class="text-sm font-semibold">아이덴티콘</div>
              </div>
            </li>

            <li class="z-ring flex items-center gap-3 rounded p-3 ring-1">
              <input
                type="radio"
                name="avatarType"
                checked={selectedType === 2}
                onchange={() => {
                  selectedType = 2
                }}
                class="accent-current"
              />
              <div class="flex items-center gap-3">
                {#if $userInfo}
                  <AvatarIcon user={$userInfo} size={60} typ={2} />
                {/if}
                <div class="text-sm font-semibold">문자 아바타</div>
              </div>
            </li>

            <li class="z-ring flex items-center gap-3 rounded p-3 ring-1">
              <input
                type="radio"
                name="avatarType"
                checked={selectedType === 3}
                onchange={() => {
                  selectedType = 3
                }}
                class="accent-current"
              />

              <div class="flex min-w-0 flex-1 items-center gap-3">
                {#if $userInfo}
                  <AvatarIcon user={$userInfo} size={60} typ={3} />
                {/if}

                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <div class="whitespace-nowrap text-sm font-semibold">그라바타</div>

                  {#if isEditingGravatar}
                    <input
                      bind:this={gravatarInput}
                      type="email"
                      bind:value={gravatarEmail}
                      placeholder="Gravatar Email"
                      class="z-base w-64 max-w-full rounded border px-3 py-2 text-sm"
                      disabled={!canEditGravatar || gravatarBusy}
                    />
                  {:else}
                    <input
                      type="text"
                      value={gravatarDisplay}
                      readonly
                      class="z-base w-64 max-w-full rounded border px-3 py-2 text-sm"
                      disabled={!canEditGravatar}
                    />
                  {/if}

                  {#if canEditGravatar}
                    {#if !isEditingGravatar}
                      <ZButton onclick={startEditGravatar}>수정</ZButton>
                    {:else}
                      <ZButton disabled={gravatarBusy || !gravatarChanged || !validateEmail(gravatarEmail)} onclick={confirmGravatar}>
                        확인
                      </ZButton>
                      <ZButton disabled={gravatarBusy} onclick={cancelEditGravatar}>취소</ZButton>
                    {/if}
                  {/if}

                  {#if gravatarError}
                    <div class="text-sm text-red-500">
                      {gravatarError}
                    </div>
                  {/if}
                </div>
              </div>
            </li>
          </ul>

          <div class="flex flex-col items-center gap-2 pt-4">
            <ZButton color="primary" disabled={!canSave} onclick={save}>저장</ZButton>

            {#if saving}
              <div class="z-muted2 text-sm">저장 중...</div>
            {:else if saveOk}
              <div class="text-sm">✅ 저장됨</div>
            {:else if saveError}
              <div class="text-sm text-red-500">
                {saveError}
              </div>
            {/if}
          </div>
        </div>
      {/if}
    </ZCard>
  {/if}
</div>
