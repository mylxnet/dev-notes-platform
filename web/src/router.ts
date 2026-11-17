import { createRouter, createWebHistory } from 'vue-router'

import HomeView from './views/HomeView.vue'
import NotesView from './views/NotesView.vue'
import NoteDetailView from './views/NoteDetailView.vue'
import ProjectsView from './views/ProjectsView.vue'
import ProjectDetailView from './views/ProjectDetailView.vue'
import AboutView from './views/AboutView.vue'

// 后台管理（/admin）在 Section 3 接入，当前只保留公开页面路由。
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/notes', name: 'notes', component: NotesView },
    { path: '/notes/:slug', name: 'note-detail', component: NoteDetailView },
    { path: '/projects', name: 'projects', component: ProjectsView },
    { path: '/projects/:slug', name: 'project-detail', component: ProjectDetailView },
    { path: '/about', name: 'about', component: AboutView },
    { path: '/:pathMatch(.*)*', redirect: '/' }
  ],
  scrollBehavior() {
    return { top: 0 }
  }
})

export default router
