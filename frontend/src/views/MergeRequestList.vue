<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="flex justify-between items-center mb-8">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">Merge Requests</h1>
        <p class="mt-2 text-gray-600">Review and manage merge requests</p>
      </div>
      <button @click="showCreateModal = true" class="btn btn-primary">
        Create Merge Request
      </button>
    </div>

    <!-- Filters -->
    <div class="card mb-6">
      <div class="flex space-x-4">
        <button 
          @click="filter = 'all'" 
          :class="filter === 'all' ? 'btn btn-primary' : 'btn btn-secondary'"
        >
          All
        </button>
        <button 
          @click="filter = 'open'" 
          :class="filter === 'open' ? 'btn btn-primary' : 'btn btn-secondary'"
        >
          Open
        </button>
        <button 
          @click="filter = 'approved'" 
          :class="filter === 'approved' ? 'btn btn-primary' : 'btn btn-secondary'"
        >
          Approved
        </button>
        <button 
          @click="filter = 'merged'" 
          :class="filter === 'merged' ? 'btn btn-primary' : 'btn btn-secondary'"
        >
          Merged
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <!-- Merge Requests List -->
    <div v-else-if="mergeRequests.length > 0" class="space-y-4">
      <div 
        v-for="mr in mergeRequests" 
        :key="mr.id"
        @click="$router.push(`/merge-requests/${mr.id}`)"
        class="card hover:shadow-md transition-shadow cursor-pointer"
      >
        <div class="flex justify-between items-start">
          <div class="flex-1">
            <div class="flex items-center space-x-2 mb-2">
              <h3 class="text-lg font-semibold text-gray-900">{{ mr.title }}</h3>
              <span 
                class="px-2 py-1 text-xs rounded"
                :class="statusClass(mr.status)"
              >
                {{ mr.status }}
              </span>
            </div>
            <p class="text-gray-600 text-sm mb-2">{{ mr.description }}</p>
            <div class="flex items-center space-x-4 text-sm text-gray-500">
              <span>by {{ mr.author }}</span>
              <span>•</span>
              <span>{{ formatDate(mr.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-12">
      <div class="text-gray-400 mb-4">
        <svg class="mx-auto h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">No merge requests</h3>
      <p class="text-gray-600 mb-4">Create your first merge request to start reviewing changes</p>
    </div>

    <!-- Create MR Modal (simplified) -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click="showCreateModal = false">
      <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4" @click.stop>
        <h2 class="text-2xl font-bold mb-4">Create Merge Request</h2>
        <p class="text-gray-600 mb-4">Merge request creation requires backend connection.</p>
        <button @click="showCreateModal = false" class="btn btn-secondary w-full">
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { mergeRequestsAPI } from '../api/client'

const loading = ref(false)
const mergeRequests = ref([])
const filter = ref('all')
const showCreateModal = ref(false)

onMounted(() => {
  fetchMergeRequests()
})

const fetchMergeRequests = async () => {
  loading.value = true
  try {
    const params = filter.value !== 'all' ? { status: filter.value } : {}
    const response = await mergeRequestsAPI.list(params)
    mergeRequests.value = response.data.merge_requests || []
  } catch (error) {
    console.error('Failed to fetch merge requests:', error)
    mergeRequests.value = []
  } finally {
    loading.value = false
  }
}

const statusClass = (status) => {
  const classes = {
    open: 'bg-yellow-100 text-yellow-800',
    approved: 'bg-green-100 text-green-800',
    merged: 'bg-blue-100 text-blue-800',
    closed: 'bg-gray-100 text-gray-800',
  }
  return classes[status] || 'bg-gray-100 text-gray-800'
}

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('en-US', { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  })
}
</script>