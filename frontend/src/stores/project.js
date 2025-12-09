import { defineStore } from 'pinia'
import { projectsAPI, branchesAPI, commitsAPI } from '../api/client'

export const useProjectStore = defineStore('project', {
  state: () => ({
    projects: [],
    currentProject: null,
    branches: [],
    currentBranch: null,
    commits: [],
    loading: false,
    error: null,
  }),

  actions: {
    async fetchProjects() {
      this.loading = true
      this.error = null
      try {
        const response = await projectsAPI.list()
        this.projects = response.data.projects || []
      } catch (error) {
        this.error = error.message
        console.error('Failed to fetch projects:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchProject(id) {
      this.loading = true
      this.error = null
      try {
        const response = await projectsAPI.get(id)
        this.currentProject = response.data
      } catch (error) {
        this.error = error.message
        console.error('Failed to fetch project:', error)
      } finally {
        this.loading = false
      }
    },

    async createProject(data) {
      this.loading = true
      this.error = null
      try {
        const response = await projectsAPI.create(data)
        this.projects.push(response.data)
        return response.data
      } catch (error) {
        this.error = error.message
        console.error('Failed to create project:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    async fetchBranches(projectId) {
      this.loading = true
      this.error = null
      try {
        const response = await branchesAPI.list(projectId)
        this.branches = response.data.branches || []
      } catch (error) {
        this.error = error.message
        console.error('Failed to fetch branches:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchBranch(id) {
      this.loading = true
      this.error = null
      try {
        const response = await branchesAPI.get(id)
        this.currentBranch = response.data
      } catch (error) {
        this.error = error.message
        console.error('Failed to fetch branch:', error)
      } finally {
        this.loading = false
      }
    },

    async createBranch(projectId, data) {
      this.loading = true
      this.error = null
      try {
        const response = await branchesAPI.create(projectId, data)
        this.branches.push(response.data)
        return response.data
      } catch (error) {
        this.error = error.message
        console.error('Failed to create branch:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    async fetchCommits(branchId, params = {}) {
      this.loading = true
      this.error = null
      try {
        const response = await commitsAPI.list(branchId, params)
        this.commits = response.data.commits || []
      } catch (error) {
        this.error = error.message
        console.error('Failed to fetch commits:', error)
      } finally {
        this.loading = false
      }
    },

    async createCommit(projectId, formData) {
      this.loading = true
      this.error = null
      try {
        const response = await commitsAPI.create(projectId, formData)
        return response.data
      } catch (error) {
        this.error = error.message
        console.error('Failed to create commit:', error)
        throw error
      } finally {
        this.loading = false
      }
    },
  },
})