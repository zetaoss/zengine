<!-- UserPage.vue -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import type { Avatar } from '@common/components/avatar/avatar'
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import { getContributions } from '@/api/action'
import { getUserInfo, getStats } from '@/api/laravel'
import ZIcon from '@common/ui/ZIcon.vue'
import ZCard from '@common/ui/ZCard.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import { mdiVectorDifference } from '@mdi/js'
import ContributionMap from './ContributionMap.vue'
import { toDate } from './util'

const route = useRoute()
const username = computed(() => route.params.username as string)

const encodedUsername = computed(() => username.value.replace(/ /g, '_'))
const userPageHref = computed(() => `/wiki/User:${encodedUsername.value}`)
const contribPageHref = computed(() => `/wiki/특수:기여/${encodedUsername.value}`)

const avatar = ref<Avatar | null>(null)
const userId = ref(0)
const editCount = ref(0)

const minDate = ref(new Date())

type StatsMap = Record<string, number>
const stats = ref<StatsMap | null>(null)

interface Contrib {
  timestamp: string
  title: string
  revid: number
}

const contribs = ref<Contrib[]>([])
const isLoading = ref(false)

const today = new Date()
today.setHours(12, 0, 0, 0)

function toAgeDate(d: Date): string {
  const diffSec = Math.floor((Date.now() - d.getTime()) / 1000)

  if (diffSec < 60) return `${diffSec}초 전`

  const min = Math.floor(diffSec / 60)
  if (min < 60) return `${min}분 전`

  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour}시간 전`

  return toDate(d)
}

function parseRegistrationDate(reg: string | undefined): Date {
  if (!reg || reg.length < 8) return new Date()
  const year = Number(reg.slice(0, 4))
  const month = Number(reg.slice(4, 6)) - 1
  const day = Number(reg.slice(6, 8))
  return new Date(year, month, day, 12, 0, 0, 0)
}

onMounted(async () => {
  isLoading.value = true
  try {
    const [userInfo, err1] = await getUserInfo(username.value)
    if (err1 || !userInfo) {
      console.error('getUserInfo error', err1)
      return
    }

    avatar.value = userInfo.avatar ?? null
    userId.value = userInfo.user_id ?? 0
    editCount.value = userInfo.user_editcount ?? 0
    minDate.value = parseRegistrationDate(userInfo.user_registration)

    const [contributions, err2] = await getContributions(userId.value)
    if (err2) {
      console.error('getContributions error', err2)
    } else {
      contribs.value = contributions ?? []
    }

    const [rawStats, err3] = await getStats(userId.value)
    if (err3) {
      console.error('getStats error', err3)
    } else {
      stats.value = rawStats ?? {}
    }
  } finally {
    isLoading.value = false
  }
})
</script>

<template>
  <div class="max-w-4xl mx-auto py-6">
    <div v-if="isLoading" class="text-center">
      <ZSpinner />
    </div>

    <div v-else>
      <ZCard class="p-6 mt-4">
        <div class="flex flex-row">
          <div class="flex items-center justify-center p-6 w-1/3">
            <AvatarIcon v-if="avatar" :avatar="avatar" :size="96" />
          </div>

          <div class="flex-1 p-6 border-l">
            <h1 class="text-xl font-semibold">
              {{ username }}
            </h1>

            <div class="space-y-1 text-sm z-muted2">
              <p>
                편집수:
                <span class="font-semibold z-text">
                  {{ editCount.toLocaleString() }}
                </span>
              </p>
              <p>
                가입일:
                <span>{{ toDate(minDate) }}</span>
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
              <td>{{ toAgeDate(new Date(c.timestamp)) }}</td>
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

      <ContributionMap v-if="stats" :stats="stats" :min-date="minDate" :max-date="today" :range-months="12" />
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
