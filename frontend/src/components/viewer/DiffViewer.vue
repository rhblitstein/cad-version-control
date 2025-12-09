<template>
  <div class="h-full flex flex-col">
    <!-- Controls -->
    <div class="bg-gray-100 p-4 rounded-t-lg flex justify-between items-center">
      <div class="flex space-x-4">
        <div class="flex items-center space-x-2">
          <input 
            type="checkbox" 
            id="sync-cameras" 
            v-model="syncCameras"
            class="rounded"
          >
          <label for="sync-cameras" class="text-sm text-gray-700">Sync Cameras</label>
        </div>
        <div class="flex items-center space-x-2">
          <input 
            type="checkbox" 
            id="show-only-changes" 
            v-model="showOnlyChanges"
            class="rounded"
          >
          <label for="show-only-changes" class="text-sm text-gray-700">Show Only Changes</label>
        </div>
      </div>
      
      <div class="text-sm text-gray-600">
        <span v-if="conflict.diff_summary?.geometry_changed" class="text-red-600 font-medium">
          ⚠️ Geometry Changed
        </span>
        <span v-else class="text-green-600 font-medium">
          ✓ No Changes
        </span>
      </div>
    </div>

    <!-- Split View -->
    <div class="flex-1 grid grid-cols-2 gap-4 p-4 bg-gray-50">
      <!-- Source Version (Left) -->
      <div class="flex flex-col">
        <div class="bg-blue-600 text-white px-4 py-2 rounded-t-lg flex justify-between items-center">
          <span class="font-semibold">Source Branch</span>
          <span class="text-xs bg-blue-700 px-2 py-1 rounded">
            {{ formatFileSize(conflict.source_version?.file_size) }}
          </span>
        </div>
        <div class="flex-1 border-2 border-blue-600 rounded-b-lg overflow-hidden">
          <StlViewer
            ref="sourceViewer"
            :file-url="conflict.source_version?.download_url || ''"
            :color="0x3b82f6"
            @camera-change="onSourceCameraChange"
          />
        </div>
      </div>

      <!-- Target Version (Right) -->
      <div class="flex flex-col">
        <div class="bg-green-600 text-white px-4 py-2 rounded-t-lg flex justify-between items-center">
          <span class="font-semibold">Target Branch</span>
          <span class="text-xs bg-green-700 px-2 py-1 rounded">
            {{ formatFileSize(conflict.target_version?.file_size) }}
          </span>
        </div>
        <div class="flex-1 border-2 border-green-600 rounded-b-lg overflow-hidden">
          <StlViewer
            ref="targetViewer"
            :file-url="conflict.target_version?.download_url || ''"
            :color="0x10b981"
            @camera-change="onTargetCameraChange"
          />
        </div>
      </div>
    </div>

    <!-- Diff Summary -->
    <div class="bg-gray-100 p-4 rounded-b-lg">
      <h3 class="text-sm font-semibold text-gray-700 mb-2">Diff Summary</h3>
      <div class="grid grid-cols-3 gap-4 text-sm">
        <div>
          <span class="text-gray-600">File Size Difference:</span>
          <span class="ml-2 font-medium" :class="sizeDiffClass">
            {{ formatSizeDiff(conflict.diff_summary?.size_diff) }}
          </span>
        </div>
        <div>
          <span class="text-gray-600">Source File:</span>
          <span class="ml-2 font-medium">{{ conflict.source_version?.filename }}</span>
        </div>
        <div>
          <span class="text-gray-600">Target File:</span>
          <span class="ml-2 font-medium">{{ conflict.target_version?.filename }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import StlViewer from './StlViewer.vue'

const props = defineProps({
  conflict: {
    type: Object,
    required: true,
  },
})

const sourceViewer = ref(null)
const targetViewer = ref(null)
const syncCameras = ref(true)
const showOnlyChanges = ref(false)

let lastCameraUpdate = 'none'

const onSourceCameraChange = (cameraData) => {
  if (syncCameras.value && lastCameraUpdate !== 'source') {
    lastCameraUpdate = 'source'
    setTimeout(() => {
      if (targetViewer.value) {
        targetViewer.value.syncCamera(cameraData.position, cameraData.target)
      }
      lastCameraUpdate = 'none'
    }, 10)
  }
}

const onTargetCameraChange = (cameraData) => {
  if (syncCameras.value && lastCameraUpdate !== 'target') {
    lastCameraUpdate = 'target'
    setTimeout(() => {
      if (sourceViewer.value) {
        sourceViewer.value.syncCamera(cameraData.position, cameraData.target)
      }
      lastCameraUpdate = 'none'
    }, 10)
  }
}

const sizeDiffClass = computed(() => {
  const diff = props.conflict.diff_summary?.size_diff || 0
  if (diff > 0) return 'text-green-600'
  if (diff < 0) return 'text-red-600'
  return 'text-gray-600'
})

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatSizeDiff = (bytes) => {
  if (!bytes) return '0 B'
  const sign = bytes > 0 ? '+' : ''
  return sign + formatFileSize(Math.abs(bytes))
}
</script>