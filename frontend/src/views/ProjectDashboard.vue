<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Back Button -->
    <button @click="$router.push('/projects')" class="mb-4 text-blue-600 hover:text-blue-800 flex items-center">
      ← Back to Projects
    </button>

    <!-- Loading State -->
    <div v-if="store.loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else-if="store.currentProject">
      <!-- Project Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900">{{ store.currentProject.name }}</h1>
        <p class="mt-2 text-gray-600">{{ store.currentProject.description }}</p>
      </div>

      <!-- Actions -->
      <div class="flex space-x-4 mb-8">
        <button @click="showBranchModal = true" class="btn btn-primary">
          Create Branch
        </button>
        <button @click="showCommitModal = true" class="btn btn-secondary">
          New Commit
        </button>
      </div>

      <!-- Branches List -->
      <div class="card">
        <h2 class="text-xl font-semibold mb-4">Branches</h2>
        
        <div v-if="store.branches.length === 0" class="text-center py-8 text-gray-500">
          No branches yet. Create your first branch to get started.
        </div>

        <div v-else class="space-y-3">
          <div 
            v-for="branch in store.branches" 
            :key="branch.id"
            @click="$router.push(`/projects/${route.params.id}/branches/${branch.id}`)"
            class="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer transition"
          >
            <div class="flex justify-between items-center">
              <div>
                <h3 class="font-semibold text-gray-900">{{ branch.name }}</h3>
                <p class="text-sm text-gray-500">
                  {{ branch.head_commit_id ? 'Has commits' : 'No commits yet' }}
                </p>
              </div>
              <div class="text-sm text-gray-400">
                Created {{ formatDate(branch.created_at) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Branch Modal -->
    <div v-if="showBranchModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click="showBranchModal = false">
      <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4" @click.stop>
        <h2 class="text-2xl font-bold mb-4">Create Branch</h2>
        <form @submit.prevent="createBranch">
          <div class="mb-4">
            <label class="label">Branch Name</label>
            <input v-model="newBranch.name" type="text" required class="input" placeholder="feature/new-design">
          </div>
          <div class="mb-6">
            <label class="label">Source Branch (optional)</label>
            <select v-model="newBranch.source_branch_id" class="input">
              <option :value="null">Create from scratch</option>
              <option v-for="branch in store.branches" :key="branch.id" :value="branch.id">
                {{ branch.name }}
              </option>
            </select>
          </div>
          <div class="flex justify-end space-x-3">
            <button type="button" @click="showBranchModal = false" class="btn btn-secondary">Cancel</button>
            <button type="submit" class="btn btn-primary">Create</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Create Commit Modal -->
    <div v-if="showCommitModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click="showCommitModal = false">
      <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4" @click.stop>
        <h2 class="text-2xl font-bold mb-4">New Commit</h2>
        <form @submit.prevent="createCommit">
          <div class="mb-4">
            <label class="label">Branch</label>
            <select v-model="newCommit.branch_id" required class="input">
              <option value="">Select a branch</option>
              <option v-for="branch in store.branches" :key="branch.id" :value="branch.id">
                {{ branch.name }}
              </option>
            </select>
          </div>
          <div class="mb-4">
            <label class="label">Commit Message</label>
            <input v-model="newCommit.message" type="text" required class="input" placeholder="Updated frame design">
          </div>
          <div class="mb-4">
            <label class="label">Author</label>
            <input v-model="newCommit.author" type="email" required class="input" placeholder="you@example.com">
          </div>
          <div class="mb-6">
            <label class="label">Upload Files</label>
            <input type="file" multiple @change="handleFileChange" class="input" accept=".stl">
          </div>
          <div class="flex justify-end space-x-3">
            <button type="button" @click="showCommitModal = false" class="btn btn-secondary">Cancel</button>
            <button type="submit" class="btn btn-primary" :disabled="uploading">
              {{ uploading ? 'Uploading...' : 'Create Commit' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useProjectStore } from '../stores/project'

const route = useRoute()
const router = useRouter()
const store = useProjectStore()

const showBranchModal = ref(false)
const showCommitModal = ref(false)
const uploading = ref(false)

const newBranch = ref({
  name: '',
  source_branch_id: null,
})

const newCommit = ref({
  branch_id: '',
  message: '',
  author: '',
  files: [],
})

onMounted(async () => {
  await store.fetchProject(route.params.id)
  await store.fetchBranches(route.params.id)
})

const createBranch = async () => {
  try {
    await store.createBranch(route.params.id, newBranch.value)
    showBranchModal.value = false
    newBranch.value = { name: '', source_branch_id: null }
  } catch (error) {
    console.error('Failed to create branch:', error)
  }
}

const handleFileChange = (event) => {
  newCommit.value.files = Array.from(event.target.files)
}

const createCommit = async () => {
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('branch_id', newCommit.value.branch_id)
    formData.append('message', newCommit.value.message)
    formData.append('author', newCommit.value.author)
    
    newCommit.value.files.forEach(file => {
      formData.append('files', file)
    })

    await store.createCommit(route.params.id, formData)
    showCommitModal.value = false
    newCommit.value = { branch_id: '', message: '', author: '', files: [] }
    
    // Refresh branches to show updated HEAD
    await store.fetchBranches(route.params.id)
  } catch (error) {
    console.error('Failed to create commit:', error)
  } finally {
    uploading.value = false
  }
}

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('en-US', { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  })
}
</script>