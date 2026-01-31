<!-- SocialJoin.vue -->
<script setup lang="ts">
import ZButton from '@common/ui/ZButton.vue'
import httpy, { HttpyError } from '@common/utils/httpy'
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

enum Status {
  Unknown = '',
  Can = 'can',
  Cannot = 'cannot',
  Checking = 'checking',
}

const router = useRouter()
const route = useRoute()

const errorMessage = ref('')
const warningMessage = ref('')
const busy = ref(false)

const token = computed(() => String(route.params.token ?? ''))

const username = ref('')
const status = ref<Status>(Status.Unknown)

watch(
  token,
  (t) => {
    if (!t) {
      errorMessage.value = 'invalid token'
      return
    }
    errorMessage.value = ''
  },
  { immediate: true },
)

function resetStatus() {
  status.value = Status.Unknown
  warningMessage.value = ''
  errorMessage.value = ''
}

function handleInvalidToken() {
  alert('세션이 만료되었습니다. 다시 로그인해 주세요.')
  router.replace('/login')
}

function getHttpStatus(err: HttpyError | null): number | undefined {
  return err?.code
}

async function checkUsername() {
  const name = username.value.trim()
  if (!name) {
    status.value = Status.Unknown
    return
  }

  status.value = Status.Checking
  warningMessage.value = ''
  errorMessage.value = ''

  const [data, err] = await httpy.post<{
    status: string
    code?: string
    message?: string
    can_create?: boolean
    name?: string
    normalized?: boolean
    messages?: string[]
  }>('/w/rest.php/social/create', {
    token: token.value,
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
    status.value = Status.Unknown
    errorMessage.value = '사용자명 확인에 실패했습니다. 잠시 후 다시 시도하세요.'
    return
  }

  if (data.status === 'success' && data.can_create === true) {
    status.value = Status.Can

    if (data.normalized && data.name && data.name !== name) {
      username.value = data.name
    }

    const warning = data.messages?.[0]
    if (warning) warningMessage.value = warning

    return
  }

  status.value = Status.Cannot
  errorMessage.value = (data.messages && data.messages[0]) || data.message || '사용불가한 사용자명입니다.'
}

async function submitJoin() {
  if (busy.value) return

  if (status.value !== Status.Can) {
    errorMessage.value = '중복 확인을 먼저 해주세요.'
    return
  }

  const name = username.value.trim()
  if (!name) return

  busy.value = true
  warningMessage.value = ''
  errorMessage.value = ''

  const [data, err] = await httpy.post<{
    status: string
    code?: string
    message?: string
    token?: string
    redirect?: string
  }>('/w/rest.php/social/create', {
    token: token.value,
    username: name,
  })

  const httpStatus = getHttpStatus(err)
  if (httpStatus === 401 || httpStatus === 403 || data?.code === 'invalid_token') {
    busy.value = false
    handleInvalidToken()
    return
  }

  if (err || !data) {
    busy.value = false
    errorMessage.value = err?.message ?? '가입 처리에 실패했습니다.'
    return
  }

  if (data.status === 'success' && data.redirect) {
    window.location.href = data.redirect
    return
  }

  errorMessage.value = data.message ?? '가입 처리에 실패했습니다.'
  busy.value = false
}
</script>

<template>
  <div class="mx-auto my-10 w-[50vw] min-w-100 rounded border z-card p-7">
    <div class="py-3 text-lg font-bold">사용자명 생성</div>
    <hr />

    <p class="py-5">사용할 사용자명을 입력하세요.</p>

    <p class="text-sm">사용자명:</p>

    <div class="flex py-2 gap-2">
      <input v-model.trim="username" aria-label="username" type="text" class="w-full p-2 border rounded" :class="status"
        placeholder="username" :disabled="busy" @input="resetStatus" @keydown.enter.prevent="checkUsername" />

      <ZButton class="whitespace-nowrap" type="button" :disabled="busy || username.trim().length < 1"
        @click="checkUsername">
        중복 확인
      </ZButton>
    </div>

    <div v-if="status === Status.Checking" class="text-sm text-gray-500">확인 중...</div>

    <div v-else-if="status === Status.Cannot" class="text-sm text-[#f008]">
      {{ errorMessage || '사용불가한 사용자명입니다.' }}
    </div>

    <div v-else-if="status === Status.Can" class="text-sm text-green-600">사용가능한 사용자명입니다.</div>

    <div v-if="warningMessage" class="bg-yellow-100 text-yellow-800 p-2 px-4 text-sm rounded my-2">
      {{ warningMessage }}
    </div>

    <div v-if="errorMessage && status !== Status.Cannot" class="bg-red-400 p-2 px-4 text-sm rounded my-2">
      {{ errorMessage }}
    </div>

    <div class="flex justify-center mt-4">
      <ZButton type="button" :disabled="busy || username.trim().length < 1 || status !== Status.Can"
        @click="submitJoin">
        가입
      </ZButton>
    </div>
  </div>
</template>

<style scoped>
@reference 'tailwindcss';

.can {
  @apply border-green-500;
}

.cannot {
  @apply border-[#f008];
}

.checking {
  @apply border-gray-400;
}
</style>
