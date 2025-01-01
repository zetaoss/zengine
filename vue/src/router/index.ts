import { createRouter, createWebHistory } from 'vue-router'
// import useAuthStore from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/views/home/TheHome.vue'),
      meta: { tab: 'wiki' },
    },
    {
      path: '/forum',
      component: () => import('@/views/forum/ForumList.vue'),
      meta: { tab: 'forum' },
    },
    {
      path: '/forum/new',
      component: () => import('@/views/forum/ForumNew.vue'),
      meta: { tab: 'forum', requiresAuth: true },
    },
    {
      path: '/forum/:id',
      component: () => import('@/views/forum/ForumList.vue'),
      meta: { tab: 'forum' },
    },
    {
      path: '/forum/:id/edit',
      component: () => import('@/views/forum/ForumEdit.vue'),
      meta: { tab: 'forum', requiresAuth: true },
    },
    {
      path: '/forum/page/:page',
      component: () => import('@/views/forum/ForumList.vue'),
      meta: { tab: 'forum' },
    },
    {
      path: '/tool',
      component: () => import('@/views/tool/TheTool.vue'),
      meta: { tab: 'tool' },
      children: [
        {
          path: 'common-report',
          component: () => import('@/views/tool/commonReport/CommonReport.vue'),
          meta: { tab: 'tool' },
        },
        {
          path: 'common-report/page/:page',
          component: () => import('@/views/tool/commonReport/CommonReport.vue'),
          meta: { tab: 'tool' },
        },
        {
          path: 'common-report/:id',
          component: () => import('@/views/tool/commonReport/CommonReportDetail.vue'),
          meta: { tab: 'tool' },
        },
        {
          path: 'write-request',
          component: () => import('@/views/tool/writeRequest/WriteRequest.vue'),
          meta: { tab: 'tool' },
        },
        {
          path: 'write-request/page/:page',
          component: () => import('@/views/tool/writeRequest/WriteRequest.vue'),
          meta: { tab: 'tool' },
        },
      ],
    },
    {
      path: '/login',
      component: () => import('@/views/auth/LoginView.vue'),
    },
    {
      path: '/logout',
      component: () => import('@/views/auth/LogoutView.vue'),
    },
    {
      path: '/user/:name',
      component: () => import('@/views/user/UserProfile.vue'),
    },
    {
      path: '/social-bridge',
      component: () => import('@/views/auth/SocialBridge.vue'),
    },
    {
      path: '/social-join/:code',
      component: () => import('@/views/auth/SocialJoin.vue'),
    },
  ],
})

export default router
