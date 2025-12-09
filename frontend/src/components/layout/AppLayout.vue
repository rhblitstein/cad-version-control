<template>
  <div class="min-h-screen flex flex-col">
    <!-- Top Navigation -->
    <nav class="bg-white border-b border-gray-200 sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <RouterLink to="/" class="flex items-center space-x-2">
              <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <span class="text-white font-bold text-lg">C</span>
              </div>
              <span class="text-xl font-bold text-gray-900">CAD Version Control</span>
            </RouterLink>
          </div>

          <div class="flex items-center space-x-4">
            <RouterLink 
              to="/projects" 
              class="px-3 py-2 rounded-md text-sm font-medium"
              :class="isActive('/projects') ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
            >
              Projects
            </RouterLink>
            <RouterLink 
              to="/merge-requests" 
              class="px-3 py-2 rounded-md text-sm font-medium"
              :class="isActive('/merge-requests') ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
            >
              Merge Requests
            </RouterLink>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="flex-1">
      <slot />
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t border-gray-200 py-4">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center text-sm text-gray-500">
        CAD Version Control System - Demo
      </div>
    </footer>
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()

const isActive = (path) => {
  return computed(() => route.path.startsWith(path))
}
</script>