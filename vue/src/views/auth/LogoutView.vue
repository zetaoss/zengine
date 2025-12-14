<!-- LogoutView.vue -->
<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import ZSpinner from '@common/ui/ZSpinner.vue'

const route = useRoute()

const getReturnto = () => {
  const r = route.query.returnto as string | undefined
  return r && r.length > 0 ? `/wiki/${r}` : '/'
}

const isRec = (v: unknown): v is Record<string, unknown> =>
  typeof v === 'object' && v !== null

async function mwLogout() {
  // CSRF token
  const tRes = await fetch('/w/api.php?action=query&meta=tokens&type=csrf&format=json', {
    credentials: 'same-origin',
  })
  if (!tRes.ok) throw new Error('Token request failed')

  const tJson = (await tRes.json()) as unknown
  if (!isRec(tJson)) throw new Error('Bad token response')

  const token = (isRec(tJson.query) && isRec(tJson.query.tokens))
    ? (tJson.query.tokens.csrftoken as string | undefined)
    : undefined
  if (!token) throw new Error('No CSRF token')

  // logout
  const body = new URLSearchParams({ action: 'logout', format: 'json', token })
  const lRes = await fetch('/w/api.php', {
    method: 'POST',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body,
  })

  if (!lRes.ok) throw new Error('Logout failed')

  try {
    const j = (await lRes.json()) as unknown
    if (isRec(j) && isRec(j.error)) throw new Error('Logout error')
  } catch {
    /* ignore */
  }
}

onMounted(async () => {
  try {
    await mwLogout()
  } catch (e) {
    console.error(e)
  } finally {
    window.location.href = getReturnto()
  }
})
</script>

<template>
  <div class="py-40 text-center">
    <ZSpinner />
    <div>Logging out...</div>
  </div>
</template>
