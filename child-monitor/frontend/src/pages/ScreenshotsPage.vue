<script setup>
import { ref, onMounted, computed } from 'vue'
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

// Preview state
const previewIndex = ref(-1)
const previewOpen = ref(false)
const preview = computed(() => screenshots.value[previewIndex.value] ?? null)

function openPreview(idx) {
  previewIndex.value = idx
  previewOpen.value = true
}
function closePreview() {
  previewOpen.value = false
}
function prevPhoto() {
  if (previewIndex.value > 0) previewIndex.value--
}
function nextPhoto() {
  if (previewIndex.value < screenshots.value.length - 1) previewIndex.value++
}

function imgURL(s) {
  return appStore.screenshotURL(s.file_path)
}

function formatTime(isoString) {
  if (!isoString) return ''
  const d = new Date(isoString)
  if (isNaN(d)) return isoString
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatDateTime(isoString) {
  if (!isoString) return ''
  const d = new Date(isoString)
  if (isNaN(d)) return isoString
  return d.toLocaleString()
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
    const id = s.id
    await DeleteScreenshot(id)
    const idx = screenshots.value.findIndex(x => x.id === id)
    if (idx !== -1) screenshots.value.splice(idx, 1)
    // Adjust preview index after deletion
    if (previewIndex.value >= screenshots.value.length) {
      previewIndex.value = screenshots.value.length - 1
    }
    if (screenshots.value.length === 0) closePreview()
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
        v-for="(s, idx) in screenshots"
        :key="s.id"
        class="col-6 col-sm-4 col-md-3 col-lg-2"
      >
        <!-- Thumbnail card with caption overlay -->
        <div
          class="screenshot-thumb cursor-pointer rounded-borders"
          @click="openPreview(idx)"
        >
          <img
            :src="imgURL(s)"
            loading="lazy"
            class="thumb-img"
            alt=""
          />
          <!-- Caption overlay at bottom of thumbnail -->
          <div class="thumb-caption">
            <div class="text-caption text-white">{{ formatTime(s.captured_at) }}</div>
            <div v-if="s.display_index > 0" class="text-caption" style="opacity:0.75;color:#ddd">
              Display {{ s.display_index }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="!loading" class="text-grey-5 text-center q-mt-xl">
      No screenshots found for the selected filters.
    </div>

    <div class="row justify-center q-mt-md" v-if="hasMore">
      <q-btn flat label="Load More" @click="load(false)" :loading="loading" />
    </div>

    <!-- Preview Dialog with prev/next navigation -->
    <q-dialog v-model="previewOpen" maximized transition-show="fade" transition-hide="fade">
      <q-card v-if="preview" class="column no-wrap" style="background:#111">
        <!-- Title bar -->
        <q-bar class="bg-dark text-white q-px-md" style="min-height:42px">
          <span class="text-caption ellipsis">{{ preview.file_name }}</span>
          <q-space />
          <span class="text-caption text-grey-5 q-mr-md">
            {{ previewIndex + 1 }} / {{ screenshots.length }}
          </span>
          <q-btn dense flat round icon="close" color="white" @click="closePreview" />
        </q-bar>

        <!-- Image area with prev/next overlay buttons -->
        <div class="col relative-position flex flex-center" style="overflow:hidden;min-height:0">
          <img
            :src="imgURL(preview)"
            style="max-width:100%; max-height:100%; object-fit:contain; display:block"
            alt=""
          />

          <!-- Previous button -->
          <q-btn
            v-if="previewIndex > 0"
            round
            color="dark"
            icon="chevron_left"
            class="preview-nav preview-nav-left"
            @click="prevPhoto"
            size="lg"
          />

          <!-- Next button -->
          <q-btn
            v-if="previewIndex < screenshots.length - 1"
            round
            color="dark"
            icon="chevron_right"
            class="preview-nav preview-nav-right"
            @click="nextPhoto"
            size="lg"
          />
        </div>

        <!-- Footer bar -->
        <q-bar class="bg-dark text-white q-px-md" style="min-height:48px">
          <div>
            <div class="text-caption text-white">{{ formatDateTime(preview.captured_at) }}</div>
            <div class="text-caption text-grey-5">{{ preview.file_path }}</div>
          </div>
          <q-space />
          <q-btn
            flat
            round
            icon="delete"
            color="negative"
            @click="remove(preview)"
            title="Delete"
          />
        </q-bar>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<style scoped>
/* Thumbnail container — fixed-ratio image box */
.screenshot-thumb {
  position: relative;
  overflow: hidden;
  background: #e0e0e0;
  aspect-ratio: 16 / 9;
}

.thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

/* Semi-transparent caption at the bottom of each thumbnail */
.thumb-caption {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 4px 6px;
  background: linear-gradient(transparent, rgba(0,0,0,0.65));
  pointer-events: none;
}

/* Prev / next overlay buttons */
.preview-nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  opacity: 0.75;
}
.preview-nav:hover { opacity: 1; }
.preview-nav-left  { left: 16px; }
.preview-nav-right { right: 16px; }
</style>
