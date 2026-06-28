import { createApp } from 'vue'
import { Quasar, Notify, Dialog } from 'quasar'
import '@quasar/extras/material-icons/material-icons.css'
import 'quasar/src/css/index.sass'
import { createPinia } from 'pinia'
import { createRouter, createWebHashHistory } from 'vue-router'

import App from './App.vue'
import DashboardPage from './pages/DashboardPage.vue'
import ScreenshotsPage from './pages/ScreenshotsPage.vue'
import ActivityLogPage from './pages/ActivityLogPage.vue'
import SettingsPage from './pages/SettingsPage.vue'
import PasswordSetupPage from './pages/PasswordSetupPage.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/',            component: DashboardPage },
    { path: '/screenshots', component: ScreenshotsPage },
    { path: '/activity',    component: ActivityLogPage },
    { path: '/settings',    component: SettingsPage },
    { path: '/setup',       component: PasswordSetupPage },
  ],
})

const app = createApp(App)
app.use(Quasar, { plugins: { Notify, Dialog } })
app.use(createPinia())
app.use(router)
app.mount('#app')
