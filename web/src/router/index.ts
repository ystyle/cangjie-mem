import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('../views/Home.vue'),
    meta: { title: '首页' }
  },
  {
    path: '/memories',
    name: 'memories',
    component: () => import('../views/Memories.vue'),
    meta: { title: '记忆列表' }
  },
  // 旧的编辑路由已废弃，功能已内联到记忆列表页面
  {
    path: '/memories/new',
    redirect: '/memories'
  },
  {
    path: '/memories/:id',
    redirect: '/memories'
  },
  {
    path: '/categories',
    name: 'categories',
    component: () => import('../views/Categories.vue'),
    meta: { title: '分类浏览' }
  },
  {
    path: '/import',
    name: 'import',
    component: () => import('../views/Import.vue'),
    meta: { title: '导入/导出' }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 设置页面标题
router.afterEach((to) => {
  const title = to.meta.title as string || 'cangjie-mem'
  document.title = `${title} - cangjie-mem`
})

export default router
