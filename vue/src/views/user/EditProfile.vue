<!-- EditProfile.vue -->
<script setup lang="ts">
import type { Avatar } from '@common/components/avatar/avatar'
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZCard from '@common/ui/ZCard.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import httpy from '@common/utils/httpy'
import { computed, nextTick,onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import useAuthStore from '@/stores/auth'

type AvatarType = 1 | 2 | 3

interface Data {
  me: Me
}

interface Me {
  avatar: Avatar
  groups: string[]
}

const route = useRoute()
const meStore = useAuthStore()

const username = computed(() => route.params.username as string)
const encodedUsername = computed(() => username.value.replace(/ /g, '_'))
const userPageHref = computed(() => `/user/${encodedUsername.value}`)

const isLoading = ref(false)
const loadError = ref<string | null>(null)

const saving = ref(false)
const saveError = ref<string | null>(null)
const saveOk = ref(false)

const me = ref<Me | null>(null)

const currentAvatar = computed(() => me.value?.avatar ?? null)
const isMe = computed(() => meStore.isLoggedIn)

const selectedType = ref<AvatarType>(1)

const gravatarEmail = ref('')
const initialGravatarEmail = ref('')

const isEditingGravatar = ref(false)
const gravatarInput = ref<HTMLInputElement | null>(null)
const gravatarBusy = ref(false)

const canEditGravatar = computed(() => selectedType.value === 3)
const canSave = computed(() => isMe.value && !saving.value)

function isAvatarType(v: unknown): v is AvatarType {
  return v === 1 || v === 2 || v === 3
}

function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())
}

const gravatarChanged = computed(
  () => gravatarEmail.value.trim() !== initialGravatarEmail.value.trim()
)

async function startEditGravatar() {
  if (!canEditGravatar.value) return
  saveOk.value = false
  saveError.value = null
  isEditingGravatar.value = true
  await nextTick()
  gravatarInput.value?.focus()
}

function cancelEditGravatar() {
  gravatarEmail.value = initialGravatarEmail.value
  isEditingGravatar.value = false
  saveOk.value = false
  saveError.value = null
}

async function confirmGravatar() {
  if (!canEditGravatar.value) return

  saveOk.value = false
  saveError.value = null

  const email = gravatarEmail.value.trim()
  if (!validateEmail(email)) {
    saveError.value = 'Gravatar 이메일 형식이 올바르지 않습니다.'
    return
  }

  gravatarBusy.value = true
  try {
    const [, verr] = await httpy.get('/api/me/gravatar/verify', { email })
    if (verr) {
      saveError.value = 'Gravatar 확인 실패'
      return
    }

    const payload: { t: AvatarType; gravatar?: string } = { t: 3, gravatar: email }
    const [, serr] = await httpy.post('/api/me/avatar', payload)
    if (serr) {
      saveError.value = '저장 실패'
      return
    }

    initialGravatarEmail.value = email
    isEditingGravatar.value = false
    saveOk.value = true

    await meStore.update()
    await load()
  } finally {
    gravatarBusy.value = false
  }
}

async function load() {
  loadError.value = null
  saveError.value = null
  saveOk.value = false

  if (!meStore.isLoggedIn) {
    await meStore.update()
  }

  const [data, err] = await httpy.get<Data>('/api/me')
  if (err) {
    loadError.value = 'failed to load profile'
    return
  }

  me.value = data.me

  const t = (data.me.avatar as unknown as { t?: unknown }).t
  selectedType.value = isAvatarType(t) ? t : 1

  const email = (data.me.avatar as unknown as { gravatar?: unknown }).gravatar
  gravatarEmail.value = typeof email === 'string' ? email : ''
  initialGravatarEmail.value = gravatarEmail.value

  isEditingGravatar.value = false
}

async function save() {
  if (!canSave.value) return

  saveError.value = null
  saveOk.value = false

  saving.value = true
  try {
    const payload: { t: AvatarType } = { t: selectedType.value }
    const [, err] = await httpy.post('/api/me/avatar', payload)
    if (err) {
      saveError.value = '저장 실패'
      return
    }

    saveOk.value = true
    await meStore.update()
    await load()
  } finally {
    saving.value = false
  }
}

watch(selectedType, () => {
  saveOk.value = false
  saveError.value = null
  isEditingGravatar.value = false
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
            <AvatarIcon v-if="currentAvatar" :avatar="currentAvatar" :size="72" />
            <div class="text-sm">
              <div class="font-semibold">{{ username }}</div>
            </div>
          </div>

          <div class="text-sm font-semibold">아바타 선택</div>

          <ul class="space-y-2">
            <li class="flex items-center gap-3 p-3 rounded ring-1 ring-[var(--z-border)]">
              <input type="radio" name="avatarType" :value="1" v-model="selectedType" class="accent-current" />
              <div class="flex items-center gap-3">
                <AvatarIcon v-if="currentAvatar" :avatar="currentAvatar" :temp-type="1" :size="60" />
                <div class="text-sm font-semibold">아이덴티콘</div>
              </div>
            </li>

            <li class="flex items-center gap-3 p-3 rounded ring-1 ring-[var(--z-border)]">
              <input type="radio" name="avatarType" :value="2" v-model="selectedType" class="accent-current" />
              <div class="flex items-center gap-3">
                <AvatarIcon v-if="currentAvatar" :avatar="currentAvatar" :temp-type="2" :size="60" />
                <div class="text-sm font-semibold">문자 아바타</div>
              </div>
            </li>

            <li class="flex items-center gap-3 p-3 rounded ring-1 ring-[var(--z-border)]">
              <input type="radio" name="avatarType" :value="3" v-model="selectedType" class="accent-current" />

              <div class="flex items-center gap-3 flex-1 min-w-0">
                <AvatarIcon v-if="currentAvatar" :avatar="currentAvatar" :temp-type="3" :size="60" />

                <div class="flex items-center gap-2 flex-wrap min-w-0">
                  <div class="text-sm font-semibold whitespace-nowrap">그라바타</div>

                  <input ref="gravatarInput" type="email" v-model="gravatarEmail" placeholder="Gravatar Email"
                    class="px-3 py-2 rounded ring-1 ring-[var(--z-border)] bg-transparent text-sm w-64 max-w-full"
                    :disabled="!canEditGravatar || !isEditingGravatar || gravatarBusy" />

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
