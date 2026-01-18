// index.ts
import { createRouter, createWebHistory } from 'vue-router'

const TheHome = () => import('@/views/home/TheHome.vue')
const ForumListPage = () => import('@/views/forum/ForumListPage.vue')
const ForumEditPage = () => import('@/views/forum/ForumEditPage.vue')
const ForumViewPage = () => import('@/views/forum/ForumViewPage.vue')

const TheTool = () => import('@/views/tool/TheTool.vue')
const CommonReport = () => import('@/views/tool/commonReport/CommonReport.vue')
const CommonReportDetail = () => import('@/views/tool/commonReport/CommonReportDetail.vue')
const FrontPlay = () => import('@/views/tool/FrontPlay.vue')
const WriteRequest = () => import('@/views/tool/writeRequest/WriteRequest.vue')
const LoginView = () => import('@/views/auth/LoginView.vue')
const LogoutView = () => import('@/views/auth/LogoutView.vue')
const SocialJoin = () => import('@/views/auth/SocialJoin.vue')
const EditProfile = () => import('@/views/user/EditProfile.vue')
const UserProfile = () => import('@/views/user/UserProfile.vue')
const TheOnelines = () => import('@/views/misc/TheOnelines.vue')

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', component: TheHome, meta: { tab: 'wiki' } },
    {
      path: '/forum',
      meta: { tab: 'forum' },
      children: [
        { path: '', component: ForumListPage },
        { path: 'new', component: ForumEditPage, meta: { requiresAuth: true } },
        { path: ':id/edit', component: ForumEditPage, meta: { requiresAuth: true } },
        { path: ':id', component: ForumViewPage },
      ],
    },
    {
      path: '/tool',
      component: TheTool,
      meta: { tab: 'tool' },
      children: [
        { path: 'common-report', component: CommonReport },
        { path: 'common-report/:id', component: CommonReportDetail },
        { path: 'write-request', component: WriteRequest },
        { path: 'frontplay', component: FrontPlay },
      ],
    },
    { path: '/login', component: LoginView },
    { path: '/logout', component: LogoutView },
    { path: '/social-join/:token', component: SocialJoin },
    { path: '/onelines', component: TheOnelines, meta: { tab: 'wiki' } },

    { path: '/user/:user_name/edit', component: EditProfile, meta: { requiresAuth: true } },
    { path: '/user/:user_name', component: UserProfile },

    // catch-all (dummy route for wiki)
    { path: '/wiki/:any(.*)', component: () => null },
  ],
})

export default router
