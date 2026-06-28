<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from './stores/app'
import AppLayout from './components/AppLayout.vue'
import { HasPassword, RequestExit, VerifyPassword, PauseMonitoring } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { useQuasar } from 'quasar'

const router = useRouter()
const appStore = useAppStore()
const $q = useQuasar()

// ── UI Lock overlay ───────────────────────────────────────────────────────────
const uiLocked = ref(false)
const lockPassword = ref('')
const lockError = ref('')
const lockLoading = ref(false)
const lockInputRef = ref(null)

// Focus the password input whenever the lock overlay appears.
watch(uiLocked, (val) => {
  if (val) {
    nextTick(() => lockInputRef.value?.focus())
  }
})

async function unlock() {
  if (!lockPassword.value) return
  lockLoading.value = true
  lockError.value = ''
  try {
    const ok = await VerifyPassword(lockPassword.value)
    if (ok) {
      uiLocked.value = false
      lockPassword.value = ''
    } else {
      lockError.value = 'Incorrect password'
    }
  } catch (e) {
    lockError.value = String(e)
  } finally {
    lockLoading.value = false
  }
}

// ── Pause confirmation dialog ─────────────────────────────────────────────────
const showPauseDialog = ref(false)
const pausePassword = ref('')
const pauseError = ref('')
const pauseLoading = ref(false)
const pauseInputRef = ref(null)

watch(showPauseDialog, (val) => {
  if (val) nextTick(() => pauseInputRef.value?.focus())
})

async function confirmPause() {
  if (!pausePassword.value) return
  pauseLoading.value = true
  pauseError.value = ''
  try {
    const ok = await VerifyPassword(pausePassword.value)
    if (ok) {
      await PauseMonitoring()
      await appStore.loadStatus()
      showPauseDialog.value = false
      pausePassword.value = ''
      $q.notify({ type: 'warning', message: 'Monitoring paused', icon: 'pause' })
    } else {
      pauseError.value = 'Incorrect password'
    }
  } catch (e) {
    pauseError.value = String(e)
  } finally {
    pauseLoading.value = false
  }
}

function cancelPause() {
  showPauseDialog.value = false
  pausePassword.value = ''
  pauseError.value = ''
}

// ── Exit dialog ───────────────────────────────────────────────────────────────
const showExitDialog = ref(false)
const exitPassword = ref('')
const exitError = ref('')
const exitInputRef = ref(null)

watch(showExitDialog, (val) => {
  if (val) nextTick(() => exitInputRef.value?.focus())
})

async function confirmExit() {
  exitError.value = ''
  try {
    await RequestExit(exitPassword.value)
  } catch (e) {
    exitError.value = String(e)
  }
}

function cancelExit() {
  showExitDialog.value = false
  exitPassword.value = ''
  exitError.value = ''
}

// ── Startup ───────────────────────────────────────────────────────────────────
onMounted(async () => {
  await appStore.loadStatus()

  const hasPassword = await HasPassword()
  if (!hasPassword) {
    router.push('/setup')
  }

  // Lock the UI whenever the window is reopened from the tray (set by beforeClose).
  EventsOn('window:lock-requested', () => {
    uiLocked.value = true
    lockPassword.value = ''
    lockError.value = ''
  })

  // Tray "Pause" — clear lock overlay so only the pause dialog is shown.
  // The pause dialog has its own password requirement.
  EventsOn('tray:pause-requested', () => {
    uiLocked.value = false
    lockPassword.value = ''
    showPauseDialog.value = true
    pausePassword.value = ''
    pauseError.value = ''
  })

  // Tray "Exit" — clear lock overlay so only the exit dialog is shown.
  // Exit is separately password-protected, lock overlay is redundant here.
  EventsOn('tray:exit-requested', () => {
    uiLocked.value = false
    lockPassword.value = ''
    showExitDialog.value = true
    exitPassword.value = ''
    exitError.value = ''
  })

  EventsOn('nav:settings', () => router.push('/settings'))
  EventsOn('monitoring:paused',  () => appStore.loadStatus())
  EventsOn('monitoring:resumed', () => appStore.loadStatus())
})
</script>

<template>
  <AppLayout />

  <!-- ── Full-screen lock overlay ──────────────────────────────────────────── -->
  <div v-if="uiLocked" class="ui-lock-overlay">
    <q-card style="width:380px" class="shadow-24">
      <q-card-section class="bg-primary text-white text-center q-pb-sm">
        <q-icon name="lock" size="36px" class="q-mb-xs" />
        <div class="text-h6">Child Monitor</div>
        <div class="text-caption" style="opacity:0.85">Enter parent password to continue</div>
      </q-card-section>

      <q-card-section class="q-pt-md">
        <q-input
          ref="lockInputRef"
          v-model="lockPassword"
          type="password"
          label="Parent Password"
          outlined
          :error="!!lockError"
          :error-message="lockError"
          @keyup.enter="unlock"
        />
      </q-card-section>

      <q-card-actions class="q-px-md q-pb-md">
        <q-btn
          color="primary"
          label="Unlock"
          class="full-width"
          size="lg"
          :loading="lockLoading"
          @click="unlock"
        />
      </q-card-actions>
    </q-card>
  </div>

  <!-- ── Pause confirmation dialog ─────────────────────────────────────────── -->
  <q-dialog v-model="showPauseDialog" @hide="cancelPause">
    <q-card style="min-width:340px">
      <q-card-section class="bg-warning text-white">
        <div class="text-h6">
          <q-icon name="pause_circle" class="q-mr-xs" />Pause Monitoring
        </div>
      </q-card-section>
      <q-card-section>
        <p class="text-body2 text-grey-7 q-mb-sm">
          Enter parent password to pause monitoring.
        </p>
        <q-input
          ref="pauseInputRef"
          v-model="pausePassword"
          type="password"
          label="Parent Password"
          outlined
          :error="!!pauseError"
          :error-message="pauseError"
          @keyup.enter="confirmPause"
        />
      </q-card-section>
      <q-card-actions align="right">
        <q-btn flat label="Cancel" @click="cancelPause" />
        <q-btn
          color="warning"
          label="Pause"
          :loading="pauseLoading"
          @click="confirmPause"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>

  <!-- ── Exit password dialog ───────────────────────────────────────────────── -->
  <q-dialog v-model="showExitDialog" @hide="cancelExit">
    <q-card style="min-width:320px">
      <q-card-section class="bg-primary text-white">
        <div class="text-h6">Exit Child Monitor</div>
      </q-card-section>
      <q-card-section>
        <q-input
          ref="exitInputRef"
          v-model="exitPassword"
          type="password"
          label="Parent Password"
          outlined
          :error="!!exitError"
          :error-message="exitError"
          @keyup.enter="confirmExit"
        />
      </q-card-section>
      <q-card-actions align="right">
        <q-btn flat label="Cancel" @click="cancelExit" />
        <q-btn color="negative" label="Exit App" @click="confirmExit" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<style>
.ui-lock-overlay {
  position: fixed;
  inset: 0;
  z-index: 99999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.82);
  backdrop-filter: blur(6px);
}
</style>
