<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { useRoute } from 'vue-router'

import http from '@/utils/http'

enum Status {
  Unknown = '',
  Can = 'can',
  Cannot = 'cannot',
}

const route = useRoute()

const username = ref('')
const status = ref(Status.Unknown)

let code = ''

async function validateCode() {
  const { data } = await http.get(`/api/auth/social/check/${code}`)
  if (!data) {
    alert('invalid code')
    window.location.href = '/'
  }
}

function changed() {
  status.value = Status.Unknown
}

async function checkUsername() {
  const { data } = await http.get('/w/api.php', {
    params: {
      action: 'query',
      list: 'users',
      ususers: username.value,
      usprop: 'cancreate',
      formatversion: '2',
      format: 'json',
      errorformat: 'html',
      errorsuselocal: 'true',
      uselang: 'ko',
    },
  })
  if (data.query.users[0].cancreate) {
    status.value = Status.Can
    return
  }
  status.value = Status.Cannot
}

async function login() {
  const { data } = await http.get(`/api/auth/social/login/${code}`, {
    params: { username: username.value },
  })
  if (data.status !== 'success') {
    alert('error on login');
    window.location.href = '/';
    return;
  }
  window.location.href = data.data;
}

async function create() {
  const { data } = await http.get(`/w/rest.php/auth/${code}`, {
    params: { username: username.value },
  })
  if (data.status !== 'success') {
    alert('error on create');
    window.location.href = '/';
    return;
  }
  login();
}

onMounted(() => {
  code = route.params.code as string;
  validateCode();
})
</script>

<template>
  <div class="bg-gray-100 mx-auto my-10 p-7 w-[50vw] min-w-[400px] rounded border">
    <div class="text-lg py-3 font-bold">
      사용자명 생성
    </div>
    <p class="py-5">
      사용할 사용자명을 입력해주세요.
    </p>
    <p class="text-sm">
      사용자명:
    </p>
    <div class="flex py-2">
      <input v-model.trim="username" aria-label="username" type="text" class="w-full p-2 border rounded" :class="status"
        placeholder="username" @input="changed">
      <button type="button" class="btn btn-secondary w-48" @click="checkUsername">
        중복 확인
      </button>
    </div>
    <div v-if="status == Status.Cannot" class="text-sm text-[#f008]">
      허용되지 않는 사용자명입니다.
    </div>
    <div class="py-8">
      <button type="button" :disabled="status != Status.Can" class="btn w-full" @click="create">
        사용자명 생성
      </button>
    </div>
  </div>
</template>

<style>
.can {
  @apply border-green-500;
}

.cannot {
  @apply border-[#f008];
}
</style>
