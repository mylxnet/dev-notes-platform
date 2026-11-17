import { createRouter, createWebHistory } from 'vue-router'

import HomeView from './views/HomeView.vue'
import NotesView from './views/NotesView.vue'
import NoteDetailView from './views/NoteDetailView.vue'
import ProjectsView from './views/ProjectsView.vue'
import ProjectDetailView from './views/ProjectDetailView.vue'
import AboutView from './views/AboutView.vue'
import AdminLoginView from './views/admin/AdminLoginView.vue'
import AdminLayout from './views/admin/AdminLayout.vue'
import AdminNotesView from './views/admin/AdminNotesView.vue'
import AdminNoteEditView from './views/admin/AdminNoteEditView.vue'
import AdminProjectsView from './views/admin/AdminProjectsView.vue'
import AdminProjectEditView from './views/admin/AdminProjectEditView.vue'
import AdminAccountView from './views/admin/AdminAccountView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
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
        { path: 'notes/new', name: 'admin-note-new', component: AdminNoteEditView },
        { path: 'notes/:id/edit', name: 'admin-note-edit', component: AdminNoteEditView },
        { path: 'projects', name: 'admin-projects', component: AdminProjectsView },
        { path: 'projects/new', name: 'admin-project-new', component: AdminProjectEditView },
        { path: 'projects/:id/edit', name: 'admin-project-edit', component: AdminProjectEditView },
        { path: 'account', name: 'admin-account', component: AdminAccountView }
      ]
    },
    { path: '/:pathMatch(.*)*', redirect: '/' }
  ],
  scrollBehavior() {
    return { top: 0 }
  }
})

export default router
