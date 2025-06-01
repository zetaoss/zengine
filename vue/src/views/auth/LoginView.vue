<script setup lang="ts">
import { ref } from 'vue'

import { mdiAccount, mdiLock } from '@mdi/js'
import { useRoute } from 'vue-router'

import Icon from '@common/ui/Icon.vue'

import GithubIcon from './icons/GithubIcon.vue'
import GoogleIcon from './icons/GoogleIcon.vue'
import doLogin from './login'

const route = useRoute()

const message = ref('')
const username = ref('')
const password = ref('')

async function login() {
  const result = await doLogin(username.value, password.value, window.location.origin)
  if (result.status === 'PASS') {
    const returnto = route.query.returnto?.toString()
    if (returnto === undefined || returnto.length < 1) {
      window.location.href = '/'
      return
    }
    if (returnto?.startsWith(`//${window.location.host}/`)) {
      window.location.href = returnto
      return
    }
    window.location.href = `/wiki/${returnto.replace(/\+/g, ' ')}`
  }
  message.value = result.message
}
</script>

<template>
  <div class="mx-auto w-[75%] max-w-xl pt-10 pb-20">
    <div class="p-6 border rounded bg-white dark:bg-slate-900">
      <div class="py-3">
        소셜 로그인
      </div>
      <div>
        <a class="btn text-sm !bg-white dark:bg-white border border-gray-300 !text-black dark:text-black hover:bg-gray-100 hover:dark:bg-gray-200 hover:text-black hover:dark:text-black"
          href="/auth/redirect/google">
          <GoogleIcon />
          <span class="ml-1">Log in with Google</span>
        </a>
        <a class="btn text-sm !bg-black dark:bg-black hover:bg-gray-700 !text-white dark:text-white hover:dark:bg-gray-800 hover:text-white hover:dark:text-white"
          href="/auth/redirect/github">
          <GithubIcon />
          <span class="ml-1.5">Log in with GitHub</span>
        </a>
      </div>
      <div class="hrbar">
        <span>또는</span>
      </div>
      <div class="py-3">
        아이디 로그인
      </div>
      <div v-if="message != ''" class="bg-red-400 p-2 px-4 text-sm rounded">
        {{ message }}
      </div>
      <div class="flex my-1">
        <span class="inline-flex items-center px-3 bg-gray-200 text-gray-500 border border-r-0 rounded-l">
          <Icon :path="mdiAccount" :size="18" />
        </span>
        <input v-model="username" aria-label="username" type="text"
          class="rounded-none rounded-r border block flex-1 min-w-0 w-full text-sm p-2.5 focus:outline-none focus:ring focus:border-blue-500 bg-white text-black"
          placeholder="사용자명" required>
      </div>
      <div class="flex my-1">
        <span class="inline-flex items-center px-3 bg-gray-200 text-gray-500 border border-r-0 rounded-l">
          <Icon :path="mdiLock" :size="18" />
        </span>
        <input v-model="password" aria-label="password" type="password"
          class="rounded-none rounded-r border block flex-1 min-w-0 w-full text-sm p-2.5 focus:outline-none focus:ring focus:border-blue-500 bg-white text-black"
          placeholder="패스워드" required autocomplete="on">
      </div>
      <button type="button" @click="login" :disabled="username === '' || password === ''" class=" my-1 rounded w-full bg-gray-800 text-white dark:text-white h-10 hover:bg-gray-700 dark:bg-[#357]
      hover:dark:bg-[#246]">
        로그인
      </button>
      <div class="py-6">
        <div class="float-left">
          <a href="/wiki/특수:비밀번호재설정">패스워드를 잊으셨나요?</a>
        </div>
        <div class="float-right">
          <a href="/wiki/특수:계정만들기">새 계정 만들기</a>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.btn {
  @apply border block my-1 p-2 rounded text-center;

  &:hover {
    text-decoration: none;
  }
}

.hrbar {
  @apply text-center relative overflow-hidden my-6 text-gray-400;
}

.hrbar span {
  @apply relative px-3;

  &:before,
  &:after {
    content: '';
    @apply absolute top-2.5 w-[999px] border-t border-[#aaa8];
  }

  &:before {
    right: 100%;
  }

  &:after {
    left: 100%;
  }
}
</style>
