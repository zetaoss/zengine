// index.ts
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/views/home/TheHome.vue'),
      meta: { tab: 'wiki' }
    },
    {
      path: '/forum',
      meta: { tab: 'forum' },
      children: [
        {
          path: '',
          component: () => import('@/views/forum/ForumListPage.vue')
        },
        {
          path: 'new',
          component: () => import('@/views/forum/ForumEditPage.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: ':id/edit',
          component: () => import('@/views/forum/ForumEditPage.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: ':id',
          component: () => import('@/views/forum/ForumViewPage.vue')
        },
      ],
    },
    {
      path: '/tool',
      component: () => import('@/views/tool/TheTool.vue'),
      meta: { tab: 'tool' },
      children: [
        {
          path: 'common-report',
          component: () => import('@/views/tool/commonReport/CommonReport.vue')
        },
        {
          path: 'common-report/:id',
          component: () => import('@/views/tool/commonReport/CommonReportDetail.vue')
        },
        {
          path: 'write-request',
          component: () => import('@/views/tool/writeRequest/WriteRequest.vue')
        },
        {
          path: 'frontplay',
          component: () => import('@/views/tool/FrontPlay.vue')
        },
      ],
    },
    {
      path: '/login',
      component: () => import('@/views/auth/LoginView.vue')
    },
    {
      path: '/logout',
      component: () => import('@/views/auth/LogoutView.vue')
    },
    {
      path: '/social-join/:token',
      component: () => import('@/views/auth/SocialJoin.vue')
    },
    {
      path: '/onelines',
      component: () => import('@/views/misc/TheOnelines.vue'),
      meta: { tab: 'wiki' }
    },
    {
      path: '/user/:user_name/edit',
      component: () => import('@/views/user/EditProfile.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/user/:user_name',
      component: () => import('@/views/user/UserProfile.vue')
    },
    // catch-all (dummy route for wiki)
    { path: '/wiki/:any(.*)', component: () => null },
  ],
})

export default router
