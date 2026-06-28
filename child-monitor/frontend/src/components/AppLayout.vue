<script setup>
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'
import { computed } from 'vue'
import PasswordDialog from './PasswordDialog.vue'
import { ref } from 'vue'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()

const showPasswordDialog = ref(false)
const passwordDialogAction = ref(null)

const navItems = [
  { label: 'Dashboard',    icon: 'dashboard',    to: '/' },
  { label: 'Screenshots',  icon: 'photo_library', to: '/screenshots' },
  { label: 'Activity Log', icon: 'timeline',      to: '/activity' },
  { label: 'Settings',     icon: 'settings',      to: '/settings' },
]

const isHidden = computed(() =>
  route.path === '/setup'
)

async function toggleMonitoring() {
  if (appStore.status.monitoringPaused) {
    await appStore.resume()
  } else {
    await appStore.pause()
  }
}
</script>

<template>
  <!-- Show bare page for password setup -->
  <router-view v-if="isHidden" />

  <q-layout view="lHh Lpr lFf" v-else>
    <!-- Top Bar -->
    <q-header elevated class="bg-primary text-white">
      <q-toolbar>
        <q-toolbar-title>
          Child Monitor
          <span class="text-caption q-ml-sm" style="opacity:0.7">v{{ appStore.status.version }}</span>
        </q-toolbar-title>

        <q-badge
          :color="appStore.status.monitoringPaused ? 'warning' : 'positive'"
          class="q-mr-md"
        >
          {{ appStore.status.monitoringPaused ? 'Paused' : 'Running' }}
        </q-badge>

        <q-btn
          flat
          :icon="appStore.status.monitoringPaused ? 'play_arrow' : 'pause'"
          :label="appStore.status.monitoringPaused ? 'Resume' : 'Pause'"
          @click="toggleMonitoring"
          size="sm"
        />
      </q-toolbar>
    </q-header>

    <!-- Left Drawer -->
    <q-drawer show-if-above :width="200" :breakpoint="500" bordered>
      <q-scroll-area class="fit">
        <q-list padding>
          <q-item
            v-for="item in navItems"
            :key="item.to"
            clickable
            :active="route.path === item.to"
            active-class="bg-blue-1 text-primary"
            @click="router.push(item.to)"
          >
            <q-item-section avatar>
              <q-icon :name="item.icon" />
            </q-item-section>
            <q-item-section>{{ item.label }}</q-item-section>
          </q-item>
        </q-list>
      </q-scroll-area>
    </q-drawer>

    <!-- Main Content -->
    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>
