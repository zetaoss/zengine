<svelte:options runes={true} />

<script lang="ts">
  import { mdiAccount, mdiGithub, mdiGoogle, mdiLock } from '@mdi/js'

  import { page } from '$app/state'
  import useAuthStore from '$lib/stores/auth'
  import CButton from '$shared/ui/CButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  import doLogin from './login'

  const auth = useAuthStore()

  let message = $state('')
  let username = $state('')
  let password = $state('')
  let loading = $state(false)

  let returnto = $derived(page.url.searchParams.get('returnto')?.trim() ?? '')
  let errorParam = $derived(page.url.searchParams.get('error')?.trim() ?? '')

  function socialErrorMessage(code: string): string {
    const byCode: Record<string, string> = {
      social_auth_failed: '소셜 인증에 실패했습니다.',
      invalid_social_id: '소셜 계정 식별값이 올바르지 않습니다.',
      social_link_failed: '소셜 계정 연결에 실패했습니다.',
    }
    return byCode[code] ?? `로그인 오류: ${code}`
  }

  function redirectAfterLogin() {
    if (returnto.startsWith(':/')) {
      const route = returnto.slice(1)
      if (!route.startsWith('//')) {
        window.location.href = route
        return
      }
    }

    if (returnto) {
      window.location.href = `/wiki/${returnto}`
      return
    }

    window.location.href = '/'
  }

  function socialHref(provider: 'google' | 'github' | 'facebook') {
    let href = `/auth/redirect/${provider}`

    if (returnto) {
      href += `?returnto=${encodeURIComponent(returnto)}`
    }

    return href
  }

  async function login() {
    if (!username || !password || loading) return

    message = ''
    loading = true

    const [clientLogin, err] = await doLogin(username, password, window.location.origin)
    if (err) {
      message = err.message
      loading = false
      return
    }

    if (!clientLogin) {
      message = 'Login failed (empty response).'
      loading = false
      return
    }

    if (clientLogin.status === 'PASS') {
      await auth.update()
      redirectAfterLogin()
      return
    }

    message = clientLogin.message || `Login failed: ${clientLogin.status}`
    loading = false
  }

  $effect(() => {
    if (message.length > 0) return
    if (!errorParam) return

    message = socialErrorMessage(errorParam)
  })
</script>

<div class="mx-auto w-[75%] max-w-xl pb-20 pt-10">
  <div class="rounded border bg-x-white p-6">
    <div class="py-3">소셜 로그인</div>

    <div class="flex flex-col gap-2">
      <CButton class="py-5 w-full" size="lg" href={socialHref('google')} rel="external" data-sveltekit-reload>
        <ZIcon path={mdiGoogle} />
        Login with Google
      </CButton>

      <CButton class="py-5 w-full" size="lg" href={socialHref('github')} rel="external" data-sveltekit-reload>
        <ZIcon path={mdiGithub} />
        Log in with GitHub
      </CButton>
    </div>

    <div class="relative my-6 overflow-hidden text-center text-x-gray-400">
      <span class="relative px-3">
        또는
        <span class="absolute right-full top-2.5 w-250 border-t border-border"></span>
        <span class="absolute left-full top-2.5 w-250 border-t border-border"></span>
      </span>
    </div>

    <div class="py-3">아이디 로그인</div>

    {#if message}
      <div class="rounded bg-x-red-400 p-2 px-4 text-sm">
        {message}
      </div>
    {/if}

    <form
      onsubmit={(e) => {
        e.preventDefault()
        void login()
      }}
    >
      <div class="my-1 flex">
        <span class="inline-flex items-center rounded-l border border-r-0 bg-x-gray-200 px-3 text-x-gray-500">
          <ZIcon path={mdiAccount} size={24} />
        </span>
        <input
          bind:value={username}
          aria-label="username"
          type="text"
          class="block w-full min-w-0 flex-1 rounded-none rounded-r border bg-x-white p-2.5 text-sm text-x-black focus:border-x-blue-500 focus:outline-none focus:ring"
          placeholder="사용자명"
          required
        />
      </div>

      <div class="my-1 flex">
        <span class="inline-flex items-center rounded-l border border-r-0 bg-x-gray-200 px-3 text-x-gray-500">
          <ZIcon path={mdiLock} size={24} />
        </span>
        <input
          bind:value={password}
          aria-label="password"
          type="password"
          class="block w-full min-w-0 flex-1 rounded-none rounded-r border bg-x-white p-2.5 text-sm text-x-black focus:border-x-blue-500 focus:outline-none focus:ring"
          placeholder="패스워드"
          required
          autocomplete="on"
        />
      </div>

      <CButton type="submit" size="lg" class="py-5 w-full h-auto" disabled={loading || username === '' || password === ''}>
        {#if loading}
          로그인 중...
        {:else}
          로그인
        {/if}
      </CButton>
    </form>

    <div class="py-6">
      <div class="float-left">
        <a class="hover:no-underline" href="/wiki/특수:비밀번호재설정" rel="external">패스워드를 잊으셨나요?</a>
      </div>
      <div class="float-right">
        <a class="hover:no-underline" href="/wiki/특수:계정만들기" rel="external">새 계정 만들기</a>
      </div>
    </div>
  </div>
</div>
