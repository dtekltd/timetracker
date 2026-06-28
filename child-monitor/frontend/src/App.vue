<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from './stores/app'
import AppLayout from './components/AppLayout.vue'
import PasswordDialog from './components/PasswordDialog.vue'
import { HasPassword, RequestExit } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { useQuasar } from 'quasar'

const router = useRouter()
const appStore = useAppStore()
const $q = useQuasar()

const showExitDialog = ref(false)
const exitPassword = ref('')
const exitError = ref('')

onMounted(async () => {
  await appStore.loadStatus()

  // Redirect to password setup if no password has been set yet.
  const hasPassword = await HasPassword()
  if (!hasPassword) {
    router.push('/setup')
  }

  // Tray requests exit -> show password dialog in frontend.
  EventsOn('tray:exit-requested', () => {
    showExitDialog.value = true
  })

  // Tray requests settings navigation.
  EventsOn('nav:settings', () => {
    router.push('/settings')
  })

  // Refresh app store on monitoring state changes.
  EventsOn('monitoring:paused',  () => appStore.loadStatus())
  EventsOn('monitoring:resumed', () => appStore.loadStatus())
})

async function confirmExit() {
  exitError.value = ''
  try {
    await RequestExit(exitPassword.value)
    // App quits; we won't reach here.
  } catch (e) {
    exitError.value = String(e)
  }
}
</script>

<template>
  <AppLayout />

  <!-- Exit password dialog (triggered from tray) -->
  <q-dialog v-model="showExitDialog">
    <q-card style="min-width:320px">
      <q-card-section class="bg-primary text-white">
        <div class="text-h6">Exit Child Monitor</div>
      </q-card-section>
      <q-card-section>
        <q-input
          v-model="exitPassword"
          type="password"
          label="Parent Password"
          autofocus
          :error="!!exitError"
          :error-message="exitError"
          @keyup.enter="confirmExit"
        />
      </q-card-section>
      <q-card-actions align="right">
        <q-btn flat label="Cancel" v-close-popup @click="exitPassword = ''; exitError = ''" />
        <q-btn color="negative" label="Exit App" @click="confirmExit" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>
