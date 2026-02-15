<svelte:options runes={true} />

<script lang="ts">
  import { page } from '$app/state'
  import ZButton from '$shared/ui/ZButton.svelte'
  import httpy, { type HttpyError } from '$shared/utils/httpy'

  const Status = {
    Unknown: '',
    Can: 'can',
    Cannot: 'cannot',
    Checking: 'checking',
  } as const
  type Status = (typeof Status)[keyof typeof Status]

  let errorMessage = $state('')
  let warningMessage = $state('')
  let busy = $state(false)

  let username = $state('')
  let status = $state<Status>(Status.Unknown)

  const token = $derived(String(page.params.token ?? ''))

  const statusClass = $derived(
    status === Status.Can
      ? 'border-green-500'
      : status === Status.Cannot
        ? 'border-[#f008]'
        : status === Status.Checking
          ? 'border-gray-400'
          : '',
  )

  $effect(() => {
    if (!token) {
      errorMessage = 'invalid token'
      return
    }
    if (errorMessage === 'invalid token') errorMessage = ''
  })

  function resetStatus() {
    status = Status.Unknown
    warningMessage = ''
    errorMessage = ''
  }

  function handleInvalidToken() {
    alert('세션이 만료되었습니다. 다시 로그인해 주세요.')
    window.location.href = '/login'
  }

  function getHttpStatus(err: HttpyError | null): number | undefined {
    return err?.code
  }

  async function checkUsername() {
    const name = username.trim()
    if (!name) {
      status = Status.Unknown
      return
    }

    status = Status.Checking
    warningMessage = ''
    errorMessage = ''

    const [data, err] = await httpy.post<{
      status: string
      code?: string
      message?: string
      can_create?: boolean
      name?: string
      normalized?: boolean
      messages?: string[]
    }>('/w/rest.php/social/create', {
      token,
      username: name,
      dryrun: true,
    })

    const httpStatus = getHttpStatus(err)
    if (httpStatus === 401 || httpStatus === 403 || data?.code === 'invalid_token') {
      handleInvalidToken()
      return
    }

    if (err || !data) {
      console.error(err)
      status = Status.Unknown
      errorMessage = '사용자명 확인에 실패했습니다. 잠시 후 다시 시도하세요.'
      return
    }

    if (data.status === 'success' && data.can_create === true) {
      status = Status.Can

      if (data.normalized && data.name && data.name !== name) {
        username = data.name
      }

      const warning = data.messages?.[0]
      if (warning) warningMessage = warning
      return
    }

    status = Status.Cannot
    errorMessage = (data.messages && data.messages[0]) || data.message || '사용불가한 사용자명입니다.'
  }

  async function submitJoin() {
    if (busy) return

    if (status !== Status.Can) {
      errorMessage = '중복 확인을 먼저 해주세요.'
      return
    }

    const name = username.trim()
    if (!name) return

    busy = true
    warningMessage = ''
    errorMessage = ''

    const [data, err] = await httpy.post<{
      status: string
      code?: string
      message?: string
      token?: string
      redirect?: string
    }>('/w/rest.php/social/create', {
      token,
      username: name,
    })

    const httpStatus = getHttpStatus(err)
    if (httpStatus === 401 || httpStatus === 403 || data?.code === 'invalid_token') {
      busy = false
      handleInvalidToken()
      return
    }

    if (err || !data) {
      busy = false
      errorMessage = err?.message ?? '가입 처리에 실패했습니다.'
      return
    }

    if (data.status === 'success' && data.redirect) {
      window.location.href = data.redirect
      return
    }

    errorMessage = data.message ?? '가입 처리에 실패했습니다.'
    busy = false
  }
</script>

<div class="z-card mx-auto my-10 w-[50vw] min-w-100 rounded border p-7">
  <div class="py-3 text-lg font-bold">사용자명 생성</div>
  <hr />

  <p class="py-5">사용할 사용자명을 입력하세요.</p>

  <p class="text-sm">사용자명:</p>

  <div class="flex gap-2 py-2">
    <input
      bind:value={username}
      aria-label="username"
      type="text"
      class={`w-full rounded border p-2 ${statusClass}`}
      placeholder="username"
      disabled={busy}
      oninput={resetStatus}
      onkeydown={(e) => {
        if (e.key === 'Enter') {
          e.preventDefault()
          void checkUsername()
        }
      }}
    />

    <ZButton class="whitespace-nowrap" type="button" disabled={busy || username.trim().length < 1} onclick={() => void checkUsername()}>
      중복 확인
    </ZButton>
  </div>

  {#if status === Status.Checking}
    <div class="text-sm text-gray-500">확인 중...</div>
  {:else if status === Status.Cannot}
    <div class="text-sm text-[#f008]">
      {errorMessage || '사용불가한 사용자명입니다.'}
    </div>
  {:else if status === Status.Can}
    <div class="text-sm text-green-600">사용가능한 사용자명입니다.</div>
  {/if}

  {#if warningMessage}
    <div class="my-2 rounded bg-yellow-100 p-2 px-4 text-sm text-yellow-800">
      {warningMessage}
    </div>
  {/if}

  {#if errorMessage && status !== Status.Cannot}
    <div class="my-2 rounded bg-red-400 p-2 px-4 text-sm">
      {errorMessage}
    </div>
  {/if}

  <div class="mt-4 flex justify-center">
    <ZButton type="button" disabled={busy || username.trim().length < 1 || status !== Status.Can} onclick={() => void submitJoin()}>
      가입
    </ZButton>
  </div>
</div>
