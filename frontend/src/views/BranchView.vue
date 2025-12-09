<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Back Button -->
    <button @click="$router.push(`/projects/${route.params.projectId}`)" class="mb-4 text-blue-600 hover:text-blue-800 flex items-center">
      ← Back to Project
    </button>

    <!-- Loading State -->
    <div v-if="store.loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else-if="store.currentBranch">
      <!-- Branch Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900">{{ store.currentBranch.name }}</h1>
        <p class="mt-2 text-gray-600">Commit History</p>
      </div>

      <!-- Commit Timeline -->
      <div class="card">
        <h2 class="text-xl font-semibold mb-6">Commits</h2>
        
        <div v-if="store.commits.length === 0" class="text-center py-8 text-gray-500">
          No commits yet on this branch.
        </div>

        <div v-else class="space-y-6">
          <div 
            v-for="(commit, index) in store.commits" 
            :key="commit.id"
            class="relative pl-8 pb-6"
            :class="{ 'border-l-2 border-gray-200': index < store.commits.length - 1 }"
          >
            <!-- Timeline Dot -->
            <div class="absolute left-0 top-0 w-4 h-4 bg-blue-600 rounded-full -translate-x-1/2"></div>
            
            <!-- Commit Info -->
            <div class="bg-gray-50 rounded-lg p-4">
              <div class="flex justify-between items-start mb-2">
                <div>
                  <h3 class="font-semibold text-gray-900">{{ commit.message }}</h3>
                  <p class="text-sm text-gray-600 mt-1">{{ commit.author }}</p>
                </div>
                <span class="text-xs text-gray-500">{{ formatDate(commit.created_at) }}</span>
              </div>
              
              <!-- File Badges -->
              <div v-if="commit.file_versions" class="flex flex-wrap gap-2 mt-3">
                <span 
                  v-for="file in commit.file_versions" 
                  :key="file.id"
                  class="px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded"
                >
                  📄 {{ file.filename }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useProjectStore } from '../stores/project'

const route = useRoute()
const store = useProjectStore()

onMounted(async () => {
  await store.fetchBranch(route.params.branchId)
  await store.fetchCommits(route.params.branchId)
})

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('en-US', { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>