<script setup>
import { ref } from 'vue'
import { VerifyPassword } from '../../wailsjs/go/main/App'

const props = defineProps({
  modelValue: Boolean,
  title: { type: String, default: 'Enter Password' },
})
const emit = defineEmits(['update:modelValue', 'success'])

const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  if (!password.value) return
  loading.value = true
  error.value = ''
  try {
    const ok = await VerifyPassword(password.value)
    if (ok) {
      emit('update:modelValue', false)
      emit('success')
      password.value = ''
    } else {
      error.value = 'Incorrect password'
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function cancel() {
  password.value = ''
  error.value = ''
  emit('update:modelValue', false)
}
</script>

<template>
  <q-dialog :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)">
    <q-card style="min-width:320px">
      <q-card-section class="bg-primary text-white">
        <div class="text-h6">{{ title }}</div>
      </q-card-section>

      <q-card-section>
        <q-input
          v-model="password"
          type="password"
          label="Password"
          autofocus
          :error="!!error"
          :error-message="error"
          @keyup.enter="submit"
        />
      </q-card-section>

      <q-card-actions align="right">
        <q-btn flat label="Cancel" @click="cancel" />
        <q-btn
          label="Confirm"
          color="primary"
          :loading="loading"
          @click="submit"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>
