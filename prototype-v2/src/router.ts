import { createRouter, createWebHashHistory } from 'vue-router'
import { isLoggedIn } from './auth'

import HomeView from './views/HomeView.vue'
import NotesView from './views/NotesView.vue'
import NoteDetailView from './views/NoteDetailView.vue'
import ProjectsView from './views/ProjectsView.vue'
import ProjectDetailView from './views/ProjectDetailView.vue'
import AboutView from './views/AboutView.vue'
import AdminLayout from './views/admin/AdminLayout.vue'
import AdminLoginView from './views/admin/AdminLoginView.vue'
import AdminNotesView from './views/admin/AdminNotesView.vue'
import AdminProjectsView from './views/admin/AdminProjectsView.vue'
import AdminAccountView from './views/admin/AdminAccountView.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/notes', name: 'notes', component: NotesView },
    { path: '/notes/:slug', name: 'note-detail', component: NoteDetailView },
    { path: '/projects', name: 'projects', component: ProjectsView },
    { path: '/projects/:slug', name: 'project-detail', component: ProjectDetailView },
    { path: '/about', name: 'about', component: AboutView },
    { path: '/admin/login', name: 'admin-login', component: AdminLoginView },
    {
      path: '/admin',
      component: AdminLayout,
      children: [
        { path: '', redirect: '/admin/notes' },
        { path: 'notes', name: 'admin-notes', component: AdminNotesView },
        { path: 'projects', name: 'admin-projects', component: AdminProjectsView },
        { path: 'account', name: 'admin-account', component: AdminAccountView }
      ]
    },
    { path: '/:pathMatch(.*)*', redirect: '/' }
  ],
  scrollBehavior() {
    return { top: 0 }
  }
})

router.beforeEach((to) => {
  if (to.path.startsWith('/admin') && to.path !== '/admin/login' && !isLoggedIn.value) {
    return { name: 'admin-login' }
  }
  return true
})

export default router
