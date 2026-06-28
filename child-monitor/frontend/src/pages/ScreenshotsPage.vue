<script setup>
import { ref, onMounted } from 'vue'
import { GetScreenshots, DeleteScreenshot } from '../../wailsjs/go/main/App'
import { useAppStore } from '../stores/app'
import { useQuasar } from 'quasar'

const $q = useQuasar()
const appStore = useAppStore()

const today = new Date().toISOString().slice(0, 10)
const filterDate = ref(today)
const filterStart = ref('')
const filterEnd = ref('')
const screenshots = ref([])
const loading = ref(false)
const offset = ref(0)
const limit = 50
const hasMore = ref(false)

const preview = ref(null)
const previewOpen = ref(false)

function openPreview(s) { preview.value = s; previewOpen.value = true }
function closePreview() { previewOpen.value = false }

function imgURL(s) {
  return appStore.status.screenshotServerURL + '/file?path=' + encodeURIComponent(s.file_path)
}

function formatTime(isoString) {
  if (!isoString) return ''
  return new Date(isoString).toLocaleString()
}

async function load(reset = false) {
  if (reset) {
    offset.value = 0
    screenshots.value = []
  }
  loading.value = true
  try {
    const results = await GetScreenshots(
      filterDate.value,
      filterStart.value,
      filterEnd.value,
      limit,
      offset.value,
    )
    const list = results || []
    screenshots.value = reset ? list : [...screenshots.value, ...list]
    hasMore.value = list.length === limit
    offset.value += list.length
  } catch (e) {
    $q.notify({ type: 'negative', message: String(e) })
  } finally {
    loading.value = false
  }
}

async function remove(s) {
  $q.dialog({
    title: 'Delete Screenshot',
    message: `Delete ${s.file_name}?`,
    cancel: true,
  }).onOk(async () => {
    await DeleteScreenshot(s.id)
    screenshots.value = screenshots.value.filter(x => x.id !== s.id)
    if (preview.value?.id === s.id) preview.value = null
    $q.notify({ type: 'positive', message: 'Screenshot deleted' })
  })
}

onMounted(() => load(true))
</script>

<template>
  <q-page padding>
    <div class="text-h5 q-mb-md">Screenshots</div>

    <!-- Filters -->
    <div class="row q-col-gutter-sm q-mb-md items-end">
      <div class="col-auto">
        <q-input v-model="filterDate" type="date" label="Date" outlined dense />
      </div>
      <div class="col-auto">
        <q-input v-model="filterStart" type="time" label="Start time" outlined dense clearable />
      </div>
      <div class="col-auto">
        <q-input v-model="filterEnd" type="time" label="End time" outlined dense clearable />
      </div>
      <div class="col-auto">
        <q-btn color="primary" label="Filter" @click="load(true)" :loading="loading" />
      </div>
    </div>

    <!-- Grid -->
    <div v-if="screenshots.length" class="row q-col-gutter-sm">
      <div
        v-for="s in screenshots"
        :key="s.id"
        class="col-6 col-sm-4 col-md-3 col-lg-2"
      >
        <q-card flat bordered class="cursor-pointer" @click="openPreview(s)">
          <q-img :src="imgURL(s)" ratio="16/9">
            <template #error>
              <div class="absolute-full flex flex-center bg-grey-3 text-grey-5 text-caption">No image</div>
            </template>
          </q-img>
          <q-card-section class="q-pa-xs">
            <div class="text-caption text-grey-7">{{ formatTime(s.captured_at) }}</div>
            <div v-if="s.display_index > 0" class="text-caption text-grey-5">
              Display {{ s.display_index }}
            </div>
          </q-card-section>
        </q-card>
      </div>
    </div>

    <div v-else-if="!loading" class="text-grey-5 text-center q-mt-xl">
      No screenshots found for the selected filters.
    </div>

    <div class="row justify-center q-mt-md" v-if="hasMore">
      <q-btn flat label="Load More" @click="load(false)" :loading="loading" />
    </div>

    <!-- Preview Dialog -->
    <q-dialog v-model="previewOpen" maximized>
      <q-card v-if="preview">
        <q-bar class="bg-primary text-white">
          <span>{{ preview.file_name }}</span>
          <q-space />
          <q-btn dense flat icon="close" @click="closePreview" />
        </q-bar>
        <q-card-section class="flex flex-center q-pa-md" style="max-height:80vh; overflow:auto">
          <q-img
            :src="imgURL(preview)"
            style="max-width:100%; max-height:70vh"
            fit="contain"
          />
        </q-card-section>
        <q-card-section>
          <div class="text-caption">{{ formatTime(preview.captured_at) }}</div>
          <div class="text-caption text-grey-6">{{ preview.file_path }}</div>
        </q-card-section>
        <q-card-actions align="right">
          <q-btn flat icon="delete" color="negative" label="Delete" @click="remove(preview)" />
          <q-btn flat label="Close" @click="closePreview" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>
