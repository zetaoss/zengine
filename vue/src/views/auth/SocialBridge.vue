<script setup lang="ts">
import { onMounted } from 'vue'

import { useRoute } from 'vue-router'

import TheSpinner from '@common/ui/Spinner.vue'

import doLogin from './login'

const route = useRoute()

async function login() {
  const username = route.query.username as string;
  const password = route.query.otp as string;
  const loginreturnurl = window.location.origin;

  const result = await doLogin(username, password, loginreturnurl)
  if (result.status === 'PASS') {
    const returnto = route.query.returnto as string;
    if (returnto === undefined || returnto.length < 1) {
      window.location.href = '/';
      return
    }
    if (returnto?.startsWith(`//${window.location.host}/`)) {
      window.location.href = returnto;
      return
    }
    window.location.href = `/wiki/${returnto.replace(/\+/g, ' ')}`;
  }
  console.error('doLogin error', result);
}

onMounted(() => {
  login();
})
</script>

<template>
  <div class="py-40 text-center">
    <TheSpinner />
    <div>Logging in...</div>
  </div>
</template>
