<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { SetPassword } from '../../wailsjs/go/main/App'
import { useAppStore } from '../stores/app'

const router = useRouter()
const appStore = useAppStore()

const password = ref('')
const confirm = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  if (password.value.length < 6) {
    error.value = 'Password must be at least 6 characters'
    return
  }
  if (password.value !== confirm.value) {
    error.value = 'Passwords do not match'
    return
  }
  loading.value = true
  try {
    await SetPassword(password.value)
    await appStore.loadStatus()
    router.push('/')
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <q-page class="flex flex-center">
    <q-card style="width:400px">
      <q-card-section class="bg-primary text-white text-center">
        <div class="text-h5">Child Monitor</div>
        <div class="text-subtitle2">Create a parent password</div>
      </q-card-section>

      <q-card-section>
        <p class="text-body2 text-grey-7 q-mb-md">
          This password protects settings and is required to exit the app.
        </p>

        <q-input
          v-model="password"
          type="password"
          label="New Password"
          class="q-mb-sm"
          hint="Minimum 6 characters"
        />
        <q-input
          v-model="confirm"
          type="password"
          label="Confirm Password"
          :error="!!error"
          :error-message="error"
          @keyup.enter="submit"
        />
      </q-card-section>

      <q-card-actions class="q-px-md q-pb-md">
        <q-btn
          label="Create Password"
          color="primary"
          class="full-width"
          :loading="loading"
          @click="submit"
        />
      </q-card-actions>
    </q-card>
  </q-page>
</template>
