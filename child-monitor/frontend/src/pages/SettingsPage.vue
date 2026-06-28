<script setup>
import { ref, onMounted } from 'vue'
import { useQuasar } from 'quasar'
import {
  GetSettings, UpdateSettings, SelectScreenshotFolder,
  OpenScreenshotFolder, OpenDataFolder,
  EnableAutoStart, DisableAutoStart,
  ChangePassword, CaptureScreenshotNow, CleanupOldScreenshots,
  GetAppVersion,
} from '../../wailsjs/go/main/App'

const captureLoading = ref(false)
const cleanupLoading = ref(false)
import PasswordDialog from '../components/PasswordDialog.vue'

const $q = useQuasar()

const authenticated = ref(false)
const showPasswordDialog = ref(true)

const settings = ref({})
const version = ref('')
const saving = ref(false)

const intervalOptions  = [1, 3, 5, 10, 15, 30].map(v => ({ label: `${v} min`, value: String(v) }))
const retentionOptions = [
  { label: '7 days',      value: '7' },
  { label: '14 days',     value: '14' },
  { label: '30 days',     value: '30' },
  { label: '60 days',     value: '60' },
  { label: '90 days',     value: '90' },
  { label: 'Never delete',value: '0' },
]
const qualityOptions   = [50, 60, 70, 80, 90, 100].map(v => ({ label: `${v}%`, value: String(v) }))

async function loadSettings() {
  settings.value = await GetSettings()
  version.value  = await GetAppVersion()
}

async function save() {
  saving.value = true
  try {
    await UpdateSettings(settings.value)
    // Sync auto-start with the toggle.
    if (settings.value.auto_start_enabled === 'true') {
      await EnableAutoStart()
    } else {
      await DisableAutoStart()
    }
    $q.notify({ type: 'positive', message: 'Settings saved' })
  } catch (e) {
    $q.notify({ type: 'negative', message: String(e) })
  } finally {
    saving.value = false
  }
}

async function pickFolder() {
  const folder = await SelectScreenshotFolder()
  if (folder) settings.value.screenshot_folder = folder
}

// Change password section
const oldPwd = ref('')
const newPwd = ref('')
const confirmPwd = ref('')
const pwdError = ref('')

async function changePassword() {
  pwdError.value = ''
  if (newPwd.value.length < 6) { pwdError.value = 'Minimum 6 characters'; return }
  if (newPwd.value !== confirmPwd.value) { pwdError.value = 'Passwords do not match'; return }
  try {
    await ChangePassword(oldPwd.value, newPwd.value)
    $q.notify({ type: 'positive', message: 'Password changed' })
    oldPwd.value = newPwd.value = confirmPwd.value = ''
  } catch (e) {
    pwdError.value = String(e)
  }
}

function onAuthenticated() {
  authenticated.value = true
  loadSettings()
}

async function openScreenshotFolder() {
  try {
    await OpenScreenshotFolder()
  } catch (e) {
    $q.notify({ type: 'negative', message: 'Cannot open folder: ' + e })
  }
}

async function openDataFolder() {
  try {
    await OpenDataFolder()
  } catch (e) {
    $q.notify({ type: 'negative', message: 'Cannot open folder: ' + e })
  }
}

async function captureNow() {
  captureLoading.value = true
  try {
    await CaptureScreenshotNow()
    $q.notify({ type: 'positive', message: 'Screenshot captured successfully' })
  } catch (e) {
    $q.notify({ type: 'negative', message: 'Capture failed: ' + e })
  } finally {
    captureLoading.value = false
  }
}

async function cleanupNow() {
  cleanupLoading.value = true
  try {
    await CleanupOldScreenshots()
    $q.notify({ type: 'positive', message: 'Old screenshots cleaned up' })
  } catch (e) {
    $q.notify({ type: 'negative', message: 'Cleanup failed: ' + e })
  } finally {
    cleanupLoading.value = false
  }
}
</script>

<template>
  <q-page padding>
    <!-- Password gate -->
    <PasswordDialog
      v-model="showPasswordDialog"
      title="Settings — Enter Password"
      @success="onAuthenticated"
    />

    <template v-if="authenticated">
      <div class="text-h5 q-mb-md">Settings</div>

      <div class="row q-col-gutter-md">
        <!-- Left column -->
        <div class="col-12 col-md-6">
          <q-card flat bordered class="q-mb-md">
            <q-card-section>
              <div class="text-subtitle1 q-mb-sm">Screenshot</div>

              <div class="row items-center q-mb-sm">
                <q-input
                  v-model="settings.screenshot_folder"
                  label="Screenshot Folder"
                  outlined
                  dense
                  readonly
                  class="col"
                />
                <q-btn flat icon="folder_open" @click="pickFolder" class="q-ml-sm" />
              </div>

              <q-select
                v-model="settings.screenshot_interval_minutes"
                :options="intervalOptions"
                label="Capture Interval"
                outlined
                dense
                emit-value
                map-options
                class="q-mb-sm"
              />

              <q-select
                v-model="settings.jpg_quality"
                :options="qualityOptions"
                label="JPG Quality"
                outlined
                dense
                emit-value
                map-options
                class="q-mb-sm"
              />

              <q-select
                v-model="settings.screenshot_retention_days"
                :options="retentionOptions"
                label="Screenshot Retention"
                outlined
                dense
                emit-value
                map-options
              />
            </q-card-section>
          </q-card>

          <q-card flat bordered class="q-mb-md">
            <q-card-section>
              <div class="text-subtitle1 q-mb-sm">Activity Tracking</div>
              <q-input
                v-model="settings.activity_sample_seconds"
                type="number"
                label="Sample interval (seconds)"
                outlined
                dense
                class="q-mb-sm"
              />
              <q-input
                v-model="settings.idle_threshold_seconds"
                type="number"
                label="Idle threshold (seconds)"
                outlined
                dense
              />
            </q-card-section>
          </q-card>
        </div>

        <!-- Right column -->
        <div class="col-12 col-md-6">
          <q-card flat bordered class="q-mb-md">
            <q-card-section>
              <div class="text-subtitle1 q-mb-sm">System</div>
              <q-toggle
                v-model="settings.auto_start_enabled"
                true-value="true"
                false-value="false"
                label="Start with Windows"
                class="q-mb-sm"
              />
              <div class="row q-col-gutter-sm">
                <div class="col-6">
                  <q-btn outline class="full-width" icon="folder" label="Screenshot Folder" @click="openScreenshotFolder" size="sm" />
                </div>
                <div class="col-6">
                  <q-btn outline class="full-width" icon="storage" label="Data Folder" @click="openDataFolder" size="sm" />
                </div>
              </div>
            </q-card-section>
          </q-card>

          <q-card flat bordered class="q-mb-md">
            <q-card-section>
              <div class="text-subtitle1 q-mb-sm">Change Password</div>
              <q-input v-model="oldPwd" type="password" label="Current Password" outlined dense class="q-mb-sm" />
              <q-input v-model="newPwd" type="password" label="New Password" outlined dense class="q-mb-sm" />
              <q-input
                v-model="confirmPwd"
                type="password"
                label="Confirm New Password"
                outlined
                dense
                :error="!!pwdError"
                :error-message="pwdError"
              />
              <q-btn color="primary" label="Change Password" @click="changePassword" class="q-mt-sm" size="sm" />
            </q-card-section>
          </q-card>

          <q-card flat bordered>
            <q-card-section>
              <div class="text-subtitle1 q-mb-sm">Actions</div>
              <div class="row q-col-gutter-sm">
                <div class="col-12">
                  <q-btn
                    outline color="primary" class="full-width" icon="photo_camera"
                    label="Capture Screenshot Now"
                    @click="captureNow"
                    :loading="captureLoading"
                    size="sm"
                  />
                </div>
                <div class="col-12">
                  <q-btn
                    outline color="warning" class="full-width" icon="cleaning_services"
                    label="Cleanup Old Screenshots Now"
                    @click="cleanupNow"
                    :loading="cleanupLoading"
                    size="sm"
                  />
                </div>
              </div>
            </q-card-section>
          </q-card>
        </div>
      </div>

      <!-- Save + version -->
      <div class="row justify-between items-center q-mt-md">
        <div class="text-caption text-grey-5">Version {{ version }}</div>
        <q-btn color="primary" label="Save Settings" icon="save" @click="save" :loading="saving" />
      </div>
    </template>
  </q-page>
</template>
