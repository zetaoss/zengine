<!-- @/views/user/EditProfile.vue -->
<script setup lang="ts">
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZCard from '@common/ui/ZCard.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import httpy from '@common/utils/httpy'
import { md5 } from 'js-md5'
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import useAuthStore from '@/stores/auth'

import { maskEmail } from './util/mask'

type AvatarType = 1 | 2 | 3

const route = useRoute()
const me = useAuthStore()

const user_name = computed(() => route.params.user_name as string)
const encodedUsername = computed(() => user_name.value.replace(/ /g, '_'))
const userPageHref = computed(() => `/user/${encodedUsername.value}`)

const isLoading = ref(false)
const loadError = ref<string | null>(null)

const saving = ref(false)
const saveError = ref<string | null>(null)
const saveOk = ref(false)

const isMe = computed(() => me.isLoggedIn && me.userInfo?.name === user_name.value)

const selectedType = ref<AvatarType>(1)

const gravatarEmail = ref('')
const gravatarHint = ref('')
const initialGravatarHint = ref('')
const verifiedGravatarHash = ref('')
const verifiedGravatarHint = ref('')
const gravatarError = ref<string | null>(null)

const isEditingGravatar = ref(false)
const gravatarInput = ref<HTMLInputElement | null>(null)
const gravatarBusy = ref(false)

const canEditGravatar = computed(() => selectedType.value === 3)
const canSave = computed(() => isMe.value && !saving.value)

function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())
}

const gravatarDisplay = computed(() => (isEditingGravatar.value ? gravatarEmail.value : gravatarHint.value))
const gravatarChanged = computed(() => gravatarEmail.value.trim() !== '')

async function startEditGravatar() {
  if (!canEditGravatar.value) return
  saveOk.value = false
  saveError.value = null
  gravatarError.value = null
  isEditingGravatar.value = true
  gravatarEmail.value = ''
  await nextTick()
  gravatarInput.value?.focus()
}

function cancelEditGravatar() {
  gravatarEmail.value = ''
  isEditingGravatar.value = false
  saveOk.value = false
  saveError.value = null
  gravatarError.value = null
}

async function confirmGravatar() {
  if (!canEditGravatar.value) return

  saveOk.value = false
  saveError.value = null
  gravatarError.value = null

  const email = gravatarEmail.value.trim()
  if (!validateEmail(email)) {
    gravatarError.value = 'Gravatar 이메일 형식이 올바르지 않습니다.'
    return
  }

  const ghash = md5(email.toLowerCase())
  const ghint = maskEmail(email)

  gravatarBusy.value = true
  try {
    const [vdata, verr] = await httpy.get<{ ok: boolean; ghash?: string }>(
      '/api/me/gravatar/verify',
      { ghash }
    )
    if (verr || !vdata?.ok) {
      gravatarError.value = 'Gravatar 확인 실패'
      return
    }

    verifiedGravatarHash.value = ghash
    verifiedGravatarHint.value = ghint
    gravatarHint.value = ghint
    initialGravatarHint.value = ghint
    gravatarEmail.value = ''
    isEditingGravatar.value = false
  } finally {
    gravatarBusy.value = false
  }
}

async function loadAvatarConfig() {
  const [cfg, cfgErr] = await httpy.get<{ t: AvatarType; ghint: string }>('/api/me/avatar')
  if (cfgErr) {
    selectedType.value = 1
    gravatarHint.value = ''
    initialGravatarHint.value = ''
    verifiedGravatarHash.value = ''
    verifiedGravatarHint.value = ''
    return
  }

  selectedType.value = cfg?.t ?? 1
  gravatarHint.value = cfg?.ghint ?? ''
  initialGravatarHint.value = gravatarHint.value
  verifiedGravatarHash.value = ''
  verifiedGravatarHint.value = ''
}

async function load() {
  loadError.value = null
  saveError.value = null
  saveOk.value = false
  gravatarError.value = null

  await me.update()
  if (!me.isLoggedIn) {
    loadError.value = 'failed to load profile'
    return
  }

  await loadAvatarConfig()
  isEditingGravatar.value = false
}

async function save() {
  if (!canSave.value) return

  saveError.value = null
  saveOk.value = false

  saving.value = true
  try {
    const payload: { t: AvatarType; ghash?: string; ghint?: string } = { t: selectedType.value }
    if (selectedType.value === 3 && verifiedGravatarHash.value !== '') {
      payload.ghash = verifiedGravatarHash.value
      payload.ghint = verifiedGravatarHint.value
    }
    const [, err] = await httpy.post('/api/me/avatar', payload)
    if (err) {
      saveError.value = '저장 실패'
      return
    }

    verifiedGravatarHash.value = ''
    verifiedGravatarHint.value = ''
    saveOk.value = true
    await me.update()
    await load()
  } finally {
    saving.value = false
  }
}

watch(selectedType, () => {
  saveOk.value = false
  saveError.value = null
  isEditingGravatar.value = false
  gravatarError.value = null
})

onMounted(() => {
  isLoading.value = true
  load().finally(() => {
    isLoading.value = false
  })
})
</script>

<template>
  <div class="max-w-4xl mx-auto py-6">
    <div v-if="isLoading" class="text-center">
      <ZSpinner />
    </div>

    <div v-else-if="loadError" class="mt-4 text-center text-sm text-red-500">
      {{ loadError }}
    </div>

    <div v-else>
      <ZCard class="p-6 mt-4">
        <template #header>
          <div class="flex items-baseline justify-between">
            <span class="text-base font-semibold">프로필 편집</span>
            <a :href="userPageHref" class="text-sm">돌아가기</a>
          </div>
        </template>

        <div v-if="!isMe" class="text-sm text-red-500">
          본인만 프로필을 수정할 수 있습니다.
        </div>

        <div v-else class="space-y-4">
          <div class="flex items-center gap-3">
            <AvatarIcon v-if="me.userInfo" :user="me.userInfo" :size="72" />
            <div class="text-sm">
              <div class="font-semibold">{{ user_name }}</div>
            </div>
          </div>

          <div class="text-sm font-semibold">아바타 선택</div>

          <ul class="space-y-2">
            <li class="flex items-center gap-3 p-3 rounded ring-1 ring-[var(--z-border)]">
              <input type="radio" name="avatarType" :value="1" v-model="selectedType" class="accent-current" />
              <div class="flex items-center gap-3">
                <AvatarIcon v-if="me.userInfo" :user="me.userInfo" :size="60" :typ="1" />
                <div class="text-sm font-semibold">아이덴티콘</div>
              </div>
            </li>

            <li class="flex items-center gap-3 p-3 rounded ring-1 ring-[var(--z-border)]">
              <input type="radio" name="avatarType" :value="2" v-model="selectedType" class="accent-current" />
              <div class="flex items-center gap-3">
                <AvatarIcon v-if="me.userInfo" :user="me.userInfo" :size="60" :typ="2" />
                <div class="text-sm font-semibold">문자 아바타</div>
              </div>
            </li>

            <li class="flex items-center gap-3 p-3 rounded ring-1 ring-[var(--z-border)]">
              <input type="radio" name="avatarType" :value="3" v-model="selectedType" class="accent-current" />

              <div class="flex items-center gap-3 flex-1 min-w-0">
                <AvatarIcon v-if="me.userInfo" :user="me.userInfo" :temp-type="3"
                  :temp-ghash="verifiedGravatarHash || null" :size="60" />
                <div class="flex items-center gap-2 flex-wrap min-w-0">
                  <div class="text-sm font-semibold whitespace-nowrap">그라바타</div>

                  <input v-if="isEditingGravatar" ref="gravatarInput" type="email" v-model="gravatarEmail"
                    placeholder="Gravatar Email"
                    class="px-3 py-2 rounded ring-1 ring-[var(--z-border)] bg-transparent text-sm w-64 max-w-full"
                    :disabled="!canEditGravatar || gravatarBusy" />

                  <input v-else type="text" :value="gravatarDisplay" readonly
                    class="px-3 py-2 rounded ring-1 ring-[var(--z-border)] bg-transparent text-sm w-64 max-w-full"
                    :disabled="!canEditGravatar" />

                  <template v-if="canEditGravatar">
                    <ZButton v-if="!isEditingGravatar" @click="startEditGravatar">
                      수정
                    </ZButton>

                    <template v-else>
                      <ZButton :disabled="gravatarBusy || !gravatarChanged || !validateEmail(gravatarEmail)"
                        @click="confirmGravatar">
                        확인
                      </ZButton>
                      <ZButton :disabled="gravatarBusy" @click="cancelEditGravatar">
                        취소
                      </ZButton>
                    </template>
                  </template>

                  <div v-if="gravatarError" class="text-sm text-red-500">
                    {{ gravatarError }}
                  </div>
                </div>
              </div>
            </li>
          </ul>

          <div class="flex flex-col items-center gap-2 pt-4">
            <ZButton color="primary" :disabled="!canSave" @click="save">
              저장
            </ZButton>

            <div v-if="saving" class="text-sm z-muted2">저장 중...</div>
            <div v-else-if="saveOk" class="text-sm">✅ 저장됨</div>
            <div v-else-if="saveError" class="text-sm text-red-500">
              {{ saveError }}
            </div>
          </div>
        </div>
      </ZCard>
    </div>
  </div>
</template>
