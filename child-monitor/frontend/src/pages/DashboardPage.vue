<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { GetDashboardData } from '../../wailsjs/go/main/App'
import { useAppStore } from '../stores/app'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const router = useRouter()
const appStore = useAppStore()

const today = new Date().toISOString().slice(0, 10)
const selectedDate = ref(today)
const data = ref(null)
const loading = ref(false)

function formatSeconds(s) {
  if (!s) return '0m'
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

async function load() {
  loading.value = true
  try {
    data.value = await GetDashboardData(selectedDate.value)
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
  EventsOn('dashboard:updated', load)
})

watch(selectedDate, load)
</script>

<template>
  <q-page padding>
    <div class="row items-center q-mb-md">
      <div class="text-h5 col">Dashboard</div>
      <q-input
        v-model="selectedDate"
        type="date"
        outlined
        dense
        style="max-width:180px"
      />
    </div>

    <q-inner-loading :showing="loading" />

    <template v-if="data">
      <!-- Summary Cards -->
      <div class="row q-col-gutter-md q-mb-lg">
        <div class="col-6 col-md-4 col-lg-2" v-for="card in [
          { label: 'Total Time',   value: formatSeconds(data.summary.total_seconds),  icon: 'schedule',       color: 'blue' },
          { label: 'Active Time',  value: formatSeconds(data.summary.active_seconds), icon: 'mouse',          color: 'green' },
          { label: 'Idle Time',    value: formatSeconds(data.summary.idle_seconds),   icon: 'coffee',         color: 'orange' },
          { label: 'Screenshots',  value: data.summary.screenshot_count,              icon: 'photo_camera',   color: 'purple' },
          { label: 'Top App',      value: data.summary.top_app || '—',               icon: 'apps',           color: 'teal' },
        ]" :key="card.label">
          <q-card flat bordered>
            <q-card-section class="text-center">
              <q-icon :name="card.icon" :color="card.color" size="28px" />
              <div class="text-h6 q-mt-xs">{{ card.value }}</div>
              <div class="text-caption text-grey-6">{{ card.label }}</div>
            </q-card-section>
          </q-card>
        </div>
      </div>

      <!-- Top Apps -->
      <div class="row q-col-gutter-md q-mb-lg">
        <div class="col-12 col-md-6">
          <q-card flat bordered>
            <q-card-section>
              <div class="text-subtitle1 q-mb-sm">Top Apps</div>
              <q-list separator>
                <q-item v-for="app in data.top_apps" :key="app.process_name" dense>
                  <q-item-section avatar>
                    <q-icon name="apps" color="grey" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label>{{ app.process_name }}</q-item-label>
                    <q-item-label caption>{{ app.last_window_title }}</q-item-label>
                  </q-item-section>
                  <q-item-section side>
                    <q-item-label>{{ formatSeconds(app.total_seconds) }}</q-item-label>
                  </q-item-section>
                </q-item>
                <q-item v-if="!data.top_apps?.length" dense>
                  <q-item-section class="text-grey-5">No data for this date</q-item-section>
                </q-item>
              </q-list>
            </q-card-section>
          </q-card>
        </div>

        <!-- Latest Screenshots -->
        <div class="col-12 col-md-6">
          <q-card flat bordered>
            <q-card-section>
              <div class="row items-center q-mb-sm">
                <div class="text-subtitle1 col">Latest Screenshots</div>
                <q-btn flat dense label="View All" @click="router.push('/screenshots')" size="sm" />
              </div>
              <div class="row q-col-gutter-xs">
                <div
                  v-for="s in data.latest_screenshots"
                  :key="s.id"
                  class="col-4"
                >
                  <div class="dash-thumb rounded-borders cursor-pointer" @click="router.push('/screenshots')">
                    <img :src="appStore.screenshotURL(s.file_path)" loading="lazy" class="dash-thumb-img" alt="" />
                    <div class="dash-thumb-caption text-caption text-white">
                      {{ new Date(s.captured_at).toLocaleTimeString([], {hour:'2-digit',minute:'2-digit'}) }}
                    </div>
                  </div>
                </div>
                <div v-if="!data.latest_screenshots?.length" class="col-12 text-grey-5 text-caption q-pa-sm">
                  No screenshots for this date
                </div>
              </div>
            </q-card-section>
          </q-card>
        </div>
      </div>

      <!-- Navigation buttons -->
      <div class="row q-col-gutter-md">
        <div class="col-auto">
          <q-btn outline icon="photo_library" label="View All Screenshots" @click="router.push('/screenshots')" />
        </div>
        <div class="col-auto">
          <q-btn outline icon="timeline" label="View Activity Log" @click="router.push('/activity')" />
        </div>
      </div>
    </template>
  </q-page>
</template>

<style scoped>
.dash-thumb {
  position: relative;
  overflow: hidden;
  background: #e0e0e0;
  aspect-ratio: 16 / 9;
}
.dash-thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.dash-thumb-caption {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 2px 4px;
  background: linear-gradient(transparent, rgba(0,0,0,0.6));
  font-size: 11px;
}
</style>
