<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="flex justify-between items-center mb-8">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">Projects</h1>
        <p class="mt-2 text-gray-600">Manage your CAD version control projects</p>
      </div>
      <button @click="showCreateModal = true" class="btn btn-primary">
        Create Project
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="store.loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <p class="mt-2 text-gray-600">Loading projects...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="store.error" class="card bg-red-50 border-red-200">
      <p class="text-red-600">{{ store.error }}</p>
    </div>

    <!-- Projects Grid -->
    <div v-else-if="store.projects.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="project in store.projects" 
        :key="project.id"
        @click="$router.push(`/projects/${project.id}`)"
        class="card hover:shadow-md transition-shadow cursor-pointer"
      >
        <h3 class="text-xl font-semibold text-gray-900 mb-2">{{ project.name }}</h3>
        <p class="text-gray-600 text-sm mb-4">{{ project.description || 'No description' }}</p>
        <div class="text-xs text-gray-500">
          Created {{ formatDate(project.created_at) }}
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-12">
      <div class="text-gray-400 mb-4">
        <svg class="mx-auto h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">No projects yet</h3>
      <p class="text-gray-600 mb-4">Get started by creating your first CAD project</p>
      <button @click="showCreateModal = true" class="btn btn-primary">
        Create Project
      </button>
    </div>

    <!-- Create Project Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click="showCreateModal = false">
      <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4" @click.stop>
        <h2 class="text-2xl font-bold mb-4">Create New Project</h2>
        <form @submit.prevent="createProject">
          <div class="mb-4">
            <label class="label">Project Name</label>
            <input v-model="newProject.name" type="text" required class="input" placeholder="Drone Frame Assembly">
          </div>
          <div class="mb-6">
            <label class="label">Description</label>
            <textarea v-model="newProject.description" rows="3" class="input" placeholder="Carbon fiber racing drone frame"></textarea>
          </div>
          <div class="flex justify-end space-x-3">
            <button type="button" @click="showCreateModal = false" class="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" class="btn btn-primary" :disabled="creating">
              {{ creating ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useProjectStore } from '../stores/project'

const store = useProjectStore()
const showCreateModal = ref(false)
const creating = ref(false)
const newProject = ref({
  name: '',
  description: '',
})

onMounted(() => {
  store.fetchProjects()
})

const createProject = async () => {
  creating.value = true
  try {
    await store.createProject(newProject.value)
    showCreateModal.value = false
    newProject.value = { name: '', description: '' }
  } catch (error) {
    console.error('Failed to create project:', error)
  } finally {
    creating.value = false
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