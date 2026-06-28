<script setup>
import { ref, onMounted, watch } from 'vue'
import { GetDailyAppUsage, GetActivityLog } from '../../wailsjs/go/main/App'

const today = new Date().toISOString().slice(0, 10)
const selectedDate = ref(today)
const search = ref('')
const appUsage = ref([])
const rawLog = ref([])
const loading = ref(false)

function formatSeconds(s) {
  if (!s) return '0m'
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

function formatTime(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleTimeString()
}

async function load() {
  loading.value = true
  try {
    const [usage, log] = await Promise.all([
      GetDailyAppUsage(selectedDate.value),
      GetActivityLog(selectedDate.value, search.value, 500, 0),
    ])
    appUsage.value = usage || []
    rawLog.value = log || []
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([selectedDate], load)

const appColumns = [
  { name: 'process_name',     label: 'App',              field: 'process_name',     align: 'left', sortable: true },
  { name: 'total_seconds',    label: 'Total',            field: r => formatSeconds(r.total_seconds),  sortable: true },
  { name: 'active_seconds',   label: 'Active',           field: r => formatSeconds(r.active_seconds), sortable: true },
  { name: 'idle_seconds',     label: 'Idle',             field: r => formatSeconds(r.idle_seconds),   sortable: true },
  { name: 'open_count',       label: 'Opens',            field: 'open_count',       align: 'right', sortable: true },
  { name: 'last_window_title',label: 'Last Window Title',field: 'last_window_title', align: 'left' },
]

const rawColumns = [
  { name: 'time',         label: 'Time',         field: r => formatTime(r.sampled_at),  align: 'left' },
  { name: 'process_name', label: 'Process',      field: 'process_name',                  align: 'left', sortable: true },
  { name: 'window_title', label: 'Window Title', field: 'window_title',                  align: 'left' },
  { name: 'is_idle',      label: 'Status',       field: r => r.is_idle ? 'Idle' : 'Active', align: 'center' },
]
</script>

<template>
  <q-page padding>
    <div class="text-h5 q-mb-md">Activity Log</div>

    <!-- Filters -->
    <div class="row q-col-gutter-sm q-mb-md items-end">
      <div class="col-auto">
        <q-input v-model="selectedDate" type="date" label="Date" outlined dense />
      </div>
      <div class="col">
        <q-input v-model="search" label="Search app / window title" outlined dense clearable @keyup.enter="load" />
      </div>
      <div class="col-auto">
        <q-btn color="primary" label="Search" @click="load" :loading="loading" />
      </div>
    </div>

    <!-- App Usage Table -->
    <div class="text-subtitle1 q-mb-sm">App Usage Summary</div>
    <q-table
      :rows="appUsage"
      :columns="appColumns"
      row-key="process_name"
      flat
      bordered
      dense
      :loading="loading"
      :pagination="{ rowsPerPage: 20 }"
      class="q-mb-lg"
    >
      <template #no-data>
        <div class="text-grey-5 q-pa-md">No app usage data for this date.</div>
      </template>
    </q-table>

    <!-- Raw Timeline -->
    <div class="text-subtitle1 q-mb-sm">Raw Activity Timeline</div>
    <q-table
      :rows="rawLog"
      :columns="rawColumns"
      row-key="id"
      flat
      bordered
      dense
      :loading="loading"
      :pagination="{ rowsPerPage: 50 }"
      virtual-scroll
    >
      <template #body-cell-is_idle="props">
        <q-td :props="props">
          <q-badge :color="props.row.is_idle ? 'orange' : 'green'" :label="props.row.is_idle ? 'Idle' : 'Active'" />
        </q-td>
      </template>
      <template #no-data>
        <div class="text-grey-5 q-pa-md">No activity data for this date.</div>
      </template>
    </q-table>
  </q-page>
</template>
