import { createRouter, createWebHistory } from 'vue-router'

const Home = () => import('@/views/home/Home.vue')
const ForumList = () => import('@/views/forum/ForumList.vue')
const ForumNew = () => import('@/views/forum/ForumNew.vue')
const ForumEdit = () => import('@/views/forum/ForumEdit.vue')
const Tool = () => import('@/views/tool/Tool.vue')
const CommonReport = () => import('@/views/tool/commonReport/CommonReport.vue')
const CommonReportDetail = () => import('@/views/tool/commonReport/CommonReportDetail.vue')
const FrontPlay = () => import('@/views/tool/FrontPlay.vue')
const WriteRequest = () => import('@/views/tool/writeRequest/WriteRequest.vue')
const LoginView = () => import('@/views/auth/LoginView.vue')
const LogoutView = () => import('@/views/auth/LogoutView.vue')
const SocialBridge = () => import('@/views/auth/SocialBridge.vue')
const SocialJoin = () => import('@/views/auth/SocialJoin.vue')
const UserPage = () => import('@/views/user/UserPage.vue')

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', component: Home, meta: { tab: 'wiki' } },
    {
      path: '/forum',
      meta: { tab: 'forum' },
      children: [
        { path: '', component: ForumList },
        { path: 'new', component: ForumNew, meta: { requiresAuth: true } },
        { path: ':id/edit', component: ForumEdit, meta: { requiresAuth: true } },
        { path: ':id', component: ForumList },
      ],
    },
    {
      path: '/tool',
      component: Tool,
      meta: { tab: 'tool' },
      children: [
        { path: 'common-report', component: CommonReport },
        { path: 'common-report/page/:page', component: CommonReport },
        { path: 'common-report/:id', component: CommonReportDetail },
        { path: 'write-request', component: WriteRequest },
        { path: 'write-request/page/:page', component: WriteRequest },
        { path: 'frontplay', component: FrontPlay },
      ],
    },
    { path: '/login', component: LoginView },
    { path: '/logout', component: LogoutView },
    { path: '/social-bridge', component: SocialBridge },
    { path: '/social-join/:code', component: SocialJoin },

    { path: '/user/:username', component: UserPage },

    // catch-all (dummy route for wiki)
    { path: '/wiki/:any', component: () => null },
  ],
})

export default router
