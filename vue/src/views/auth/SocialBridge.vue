<script setup lang="ts">
import ZSpinner from '@common/ui/ZSpinner.vue'
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'

import doLogin from './login'

const route = useRoute()

function redirect() {
  const returnto = route.query.returnto as string | undefined

  if (returnto && returnto.length > 0) {
    window.location.href = `/wiki/${returnto.replace(/\+/g, ' ')}`
  } else {
    window.location.href = '/'
  }
}

async function login() {
  try {
    const username = route.query.username as string
    const password = route.query.otp as string
    const loginreturnurl = window.location.origin

    const [clientLogin, err] = await doLogin(username, password, loginreturnurl)
    if (err) {
      console.error(err)
      return
    }

    if (clientLogin?.status !== 'PASS') {
      console.error('not pass', clientLogin)
    }
  } catch (e) {
    console.error('unexpected', e)
  } finally {
    redirect()
  }
}

onMounted(login)
</script>

<template>
  <div class="py-40 text-center">
    <ZSpinner />
    <div>Logging in...</div>
  </div>
</template>
