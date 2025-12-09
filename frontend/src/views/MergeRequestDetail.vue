<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Back Button -->
    <button @click="$router.push('/merge-requests')" class="mb-4 text-blue-600 hover:text-blue-800 flex items-center">
      ← Back to Merge Requests
    </button>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else-if="mergeRequest">
      <!-- MR Header -->
      <div class="card mb-6">
        <div class="flex justify-between items-start mb-4">
          <div class="flex-1">
            <h1 class="text-2xl font-bold text-gray-900 mb-2">{{ mergeRequest.title }}</h1>
            <p class="text-gray-600">{{ mergeRequest.description }}</p>
          </div>
          <span 
            class="px-3 py-1 text-sm rounded"
            :class="statusClass(mergeRequest.status)"
          >
            {{ mergeRequest.status }}
          </span>
        </div>
        
        <div class="flex items-center space-x-4 text-sm text-gray-500">
          <span>by {{ mergeRequest.author }}</span>
          <span>•</span>
          <span>{{ formatDate(mergeRequest.created_at) }}</span>
        </div>

        <!-- Actions -->
        <div v-if="mergeRequest.status === 'open'" class="flex space-x-3 mt-4">
          <button @click="approve" class="btn btn-primary">Approve</button>
          <button @click="merge" class="btn btn-secondary" :disabled="hasUnresolvedConflicts">
            {{ hasUnresolvedConflicts ? 'Has Conflicts' : 'Merge' }}
          </button>
        </div>
      </div>

      <!-- Tabs -->
      <div class="card mb-6">
        <div class="flex space-x-6 border-b border-gray-200">
          <button 
            @click="activeTab = 'conflicts'" 
            :class="tabClass('conflicts')"
          >
            Conflicts ({{ conflicts.length }})
          </button>
          <button 
            @click="activeTab = 'discussion'" 
            :class="tabClass('discussion')"
          >
            Discussion ({{ comments.length }})
          </button>
          <button 
            @click="activeTab = 'approvals'" 
            :class="tabClass('approvals')"
          >
            Approvals ({{ approvals.length }})
          </button>
        </div>
      </div>

      <!-- Conflicts Tab -->
      <div v-if="activeTab === 'conflicts'" class="space-y-4">
        <div v-if="conflicts.length === 0" class="card text-center py-8 text-gray-500">
          ✅ No conflicts! Ready to merge.
        </div>
        
        <div v-else>
          <div 
            v-for="conflict in conflicts" 
            :key="conflict.id"
            class="card"
          >
            <div class="flex justify-between items-start mb-3">
              <div>
                <h3 class="font-semibold text-gray-900">File Conflict</h3>
                <p class="text-sm text-gray-600">Different versions detected</p>
              </div>
              <span 
                class="px-2 py-1 text-xs rounded"
                :class="conflict.status === 'resolved' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
              >
                {{ conflict.status }}
              </span>
            </div>
            
            <div class="flex space-x-3">
              <button @click="viewDiff(conflict.id)" class="btn btn-primary">
                View 3D Diff
              </button>
              <button 
                v-if="conflict.status === 'unresolved'"
                @click="resolveConflict(conflict.id)" 
                class="btn btn-secondary"
              >
                Mark Resolved
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Discussion Tab -->
      <div v-if="activeTab === 'discussion'" class="space-y-4">
        <!-- Comments List -->
        <div v-for="comment in comments" :key="comment.id" class="card">
          <div class="flex items-start space-x-3">
            <div class="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white text-sm">
              {{ comment.author[0].toUpperCase() }}
            </div>
            <div class="flex-1">
              <div class="flex items-center space-x-2 mb-1">
                <span class="font-semibold text-gray-900">{{ comment.author }}</span>
                <span class="text-xs text-gray-500">{{ formatDate(comment.created_at) }}</span>
              </div>
              <p class="text-gray-700">{{ comment.content }}</p>
            </div>
          </div>
        </div>

        <!-- Add Comment -->
        <div class="card">
          <textarea 
            v-model="newComment" 
            rows="3" 
            class="input mb-3" 
            placeholder="Add a comment..."
          ></textarea>
          <button @click="addComment" class="btn btn-primary">
            Post Comment
          </button>
        </div>
      </div>

      <!-- Approvals Tab -->
      <div v-if="activeTab === 'approvals'" class="space-y-4">
        <div v-if="approvals.length === 0" class="card text-center py-8 text-gray-500">
          No approvals yet
        </div>
        
        <div v-else>
          <div v-for="approval in approvals" :key="approval.id" class="card">
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-green-600 rounded-full flex items-center justify-center text-white">
                ✓
              </div>
              <div class="flex-1">
                <span class="font-semibold text-gray-900">{{ approval.approver }}</span>
                <span class="text-sm text-gray-500 ml-2">approved {{ formatDate(approval.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 3D Diff Modal -->
    <div v-if="showDiffModal" class="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50" @click="showDiffModal = false">
      <div class="bg-white rounded-lg p-6 max-w-6xl w-full mx-4 h-5/6" @click.stop>
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-2xl font-bold">3D Diff Viewer</h2>
          <button @click="showDiffModal = false" class="text-gray-500 hover:text-gray-700">✕</button>
        </div>
        
        <DiffViewer v-if="currentDiff" :conflict="currentDiff" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { mergeRequestsAPI, conflictsAPI } from '../api/client'
import DiffViewer from '../components/viewer/DiffViewer.vue'

const route = useRoute()
const loading = ref(false)
const mergeRequest = ref(null)
const conflicts = ref([])
const comments = ref([])
const approvals = ref([])
const activeTab = ref('conflicts')
const newComment = ref('')
const showDiffModal = ref(false)
const currentDiff = ref(null)

onMounted(() => {
  fetchMergeRequest()
})

const fetchMergeRequest = async () => {
  loading.value = true
  try {
    const response = await mergeRequestsAPI.get(route.params.id)
    mergeRequest.value = response.data.merge_request
    conflicts.value = response.data.conflicts || []
    comments.value = response.data.comments || []
    approvals.value = response.data.approvals || []
  } catch (error) {
    console.error('Failed to fetch merge request:', error)
  } finally {
    loading.value = false
  }
}

const hasUnresolvedConflicts = computed(() => {
  return conflicts.value.some(c => c.status === 'unresolved')
})

const approve = async () => {
  try {
    await mergeRequestsAPI.approve(route.params.id, { 
      approver: 'you@example.com' 
    })
    await fetchMergeRequest()
  } catch (error) {
    console.error('Failed to approve:', error)
  }
}

const merge = async () => {
  try {
    await mergeRequestsAPI.merge(route.params.id)
    await fetchMergeRequest()
  } catch (error) {
    console.error('Failed to merge:', error)
  }
}

const viewDiff = async (conflictId) => {
  try {
    const response = await conflictsAPI.getDiff(conflictId)
    currentDiff.value = response.data
    showDiffModal.value = true
  } catch (error) {
    console.error('Failed to get diff:', error)
  }
}

const resolveConflict = async (conflictId) => {
  try {
    await conflictsAPI.resolve(conflictId, {
      resolution_notes: 'Manually resolved'
    })
    await fetchMergeRequest()
  } catch (error) {
    console.error('Failed to resolve conflict:', error)
  }
}

const addComment = async () => {
  if (!newComment.value.trim()) return
  
  try {
    await mergeRequestsAPI.addComment(route.params.id, {
      author: 'you@example.com',
      content: newComment.value
    })
    newComment.value = ''
    await fetchMergeRequest()
  } catch (error) {
    console.error('Failed to add comment:', error)
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

const tabClass = (tab) => {
  return activeTab.value === tab 
    ? 'px-4 py-2 border-b-2 border-blue-600 text-blue-600 font-medium'
    : 'px-4 py-2 text-gray-600 hover:text-gray-900'
}

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