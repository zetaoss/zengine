<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { page } from '$app/state'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy, { HttpyError } from '$shared/utils/httpy'

  function getReturnto() {
    const r = (page.url.searchParams.get('returnto') ?? '').trim()
    return r.length > 0 ? `/wiki/${r}` : '/'
  }

  async function getCsrfToken(): Promise<[string, HttpyError | null]> {
    const [data, err] = await httpy.get<{
      query?: { tokens?: { csrftoken?: string } }
    }>('/w/api.php', {
      action: 'query',
      meta: 'tokens',
      type: 'csrf',
      format: 'json',
    })
    if (err) return ['', err]

    const token = data.query?.tokens?.csrftoken ?? ''
    if (!token) return ['', new HttpyError(0, 'BAD_DATA', 'No CSRF token')]
    return [token, null]
  }

  async function mwLogout() {
    const [token, tokenErr] = await getCsrfToken()
    if (tokenErr) throw tokenErr

    const form = new FormData()
    form.append('action', 'logout')
    form.append('format', 'json')
    form.append('token', token)

    const [, logoutErr] = await httpy.post<Record<string, unknown>>('/w/api.php', form)
    if (logoutErr) throw logoutErr
  }

  onMount(async () => {
    try {
      await mwLogout()
    } catch (e) {
      console.error(e)
    } finally {
      window.location.href = getReturnto()
    }
  })
</script>

<div class="py-40 text-center">
  <div class="inline-flex items-center">
    <ZSpinner />
    <div>Logging out...</div>
  </div>
</div>
