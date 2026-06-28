import { defineStore } from 'pinia'
import { ref } from 'vue'
import { GetAppStatus, PauseMonitoring, ResumeMonitoring } from '../../wailsjs/go/main/App'

export const useAppStore = defineStore('app', () => {
  const status = ref({
    version: '',
    monitoringPaused: false,
    autoStartEnabled: false,
    screenshotFolder: '',
    screenshotServerURL: '',
  })

  async function loadStatus() {
    try {
      status.value = await GetAppStatus()
    } catch (e) {
      console.error('loadStatus', e)
    }
  }

  async function pause() {
    await PauseMonitoring()
    status.value.monitoringPaused = true
  }

  async function resume() {
    await ResumeMonitoring()
    status.value.monitoringPaused = false
  }

  return { status, loadStatus, pause, resume }
})
