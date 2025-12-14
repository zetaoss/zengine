<!-- UserPage.vue -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import type { Avatar } from '@common/components/avatar/avatar'
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import ZCard from '@common/ui/ZCard.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import { mdiVectorDifference } from '@mdi/js'
import ContributionMap from './ContributionMap.vue'
import { useDateFormat } from '@vueuse/core'

import httpy from '@common/utils/httpy'
import useAuthStore from '@/stores/auth'

const route = useRoute()
const username = computed(() => route.params.username as string)

const encodedUsername = computed(() => username.value.replace(/ /g, '_'))
const userPageHref = computed(() => `/wiki/User:${encodedUsername.value}`)
const contribPageHref = computed(() => `/wiki/특수:기여/${encodedUsername.value}`)

const me = useAuthStore()
const userId = ref(0)
const avatar = ref<Avatar | null>(null)
const editCount = ref(0)

const isMe = computed(() => me.isLoggedIn && me.userData.avatar.id === userId.value)

const minDate = ref(new Date())

type StatsMap = Record<string, number>
const stats = ref<StatsMap | null>(null)

interface Contribution {
  timestamp: string
  title: string
  revid: number
}

interface UserContribsResponse {
  usercontribs?: Contribution[]
}

interface ActionQueryResponse<T> {
  query?: T
  error?: { info?: string }
}

interface UserInfo {
  user_id: number
  user_name: string
  user_registration: string
  user_editcount: number
  avatar: Avatar
}

const contribs = ref<Contribution[]>([])

const isLoading = ref(false)
const loadError = ref<string | null>(null)

const today = new Date()
today.setHours(12, 0, 0, 0)

function agoDate(d: Date): string {
  const diffSec = Math.floor((Date.now() - d.getTime()) / 1000)
  if (diffSec < 60) return `${diffSec}초 전`
  const min = Math.floor(diffSec / 60)
  if (min < 60) return `${min}분 전`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour}시간 전`
  return useDateFormat(d, 'YYYY-MM-DD').value
}

function parseRegistrationDate(reg: string): Date {
  if (!reg || reg.length < 8) return new Date()
  const [y, m, d] = [reg.slice(0, 4), reg.slice(4, 6), reg.slice(6, 8)].map(Number)
  return new Date(y, m - 1, d, 12)
}

async function fetchOrSetError<T>(
  resource: string,
  promise: ReturnType<typeof httpy.get<T>>,
): Promise<T | null> {
  const [data, err] = await promise
  if (err) {
    loadError.value = `failed to get ${resource}`
    return null
  }
  return data
}

async function load() {
  loadError.value = null

  if (!me.isLoggedIn) await me.update()

  const userInfo = await fetchOrSetError<UserInfo>(
    'UserInfo',
    httpy.get(`/api/user/${encodeURIComponent(username.value)}`),
  )
  if (!userInfo) return

  avatar.value = userInfo.avatar
  userId.value = userInfo.user_id
  editCount.value = userInfo.user_editcount
  minDate.value = parseRegistrationDate(userInfo.user_registration)

  const contribRes = await fetchOrSetError<ActionQueryResponse<UserContribsResponse>>(
    'UserContribs',
    httpy.get('/w/api.php', {
      action: 'query',
      format: 'json',
      list: 'usercontribs',
      ucuser: username.value,
      uclimit: '10',
      ucprop: 'ids|title|timestamp',
    }),
  )
  if (!contribRes) return
  contribs.value = contribRes.query?.usercontribs ?? []

  const rawStats = await fetchOrSetError<StatsMap>(
    'StatsMap',
    httpy.get(`/api/user/${userId.value}/stats`),
  )
  if (!rawStats) return

  stats.value = rawStats
}

onMounted(() => {
  isLoading.value = true
  loadError.value = null

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
        <div class="flex flex-row">
          <div class="flex items-center justify-center p-6 w-1/3">
            <AvatarIcon v-if="avatar" :avatar="avatar" :size="96" />
          </div>

          <div class="flex-1 p-6 border-l">
            <div class="flex items-baseline gap-2">
              <h1 class="text-xl font-semibold">{{ username }}</h1>
              <a v-if="isMe" :href="`/user/${encodedUsername}/edit`" class="text-xs ml-1">
                Edit Profile
              </a>
            </div>

            <div class="space-y-1 text-sm z-muted2 mt-2">
              <p>
                편집수:
                <span class="font-semibold z-text">
                  {{ editCount.toLocaleString() }}
                </span>
              </p>
              <p>
                가입일:
                {{ useDateFormat(minDate, 'YYYY-MM-DD') }}
              </p>
              <p>
                <a :href="userPageHref">사용자 문서 바로가기</a>
              </p>
            </div>
          </div>
        </div>
      </ZCard>

      <ZCard class="p-6 mt-4">
        <template #header>
          <a :href="contribPageHref">최근 편집</a>
        </template>

        <table class="mytable w-full z-muted2">
          <thead>
            <tr class="border-b">
              <th>일시</th>
              <th>문서</th>
              <th>차이</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in contribs" :key="c.revid" class="border-b">
              <td>{{ agoDate(new Date(c.timestamp)) }}</td>
              <td>
                <a :href="`/wiki/${encodeURIComponent(c.title.replace(/ /g, '_'))}`">
                  {{ c.title }}
                </a>
              </td>
              <td>
                <a :href="`/w/index.php?title=${encodeURIComponent(c.title)}&diff=prev&oldid=${c.revid}`">
                  <ZIcon :path="mdiVectorDifference" />
                </a>
              </td>
            </tr>
          </tbody>
        </table>
      </ZCard>

      <ContributionMap v-if="stats" :stats="stats" :min-date="minDate" :max-date="today" />
    </div>
  </div>
</template>

<style scoped>
.mytable td,
.mytable th {
  text-align: left;
  padding: 2px 0;
}
</style>
