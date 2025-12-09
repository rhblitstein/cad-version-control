<template>
  <div ref="container" class="w-full h-full relative bg-gray-900 rounded-lg overflow-hidden">
    <!-- Loading State -->
    <div v-if="loading" class="absolute inset-0 flex items-center justify-center bg-gray-900 bg-opacity-75">
      <div class="text-white">Loading 3D model...</div>
    </div>
    
    <!-- Controls -->
    <div class="absolute top-4 right-4 flex flex-col space-y-2">
      <button 
        @click="toggleWireframe" 
        class="px-3 py-2 bg-white bg-opacity-90 rounded shadow hover:bg-opacity-100 text-sm"
      >
        {{ wireframe ? 'Solid' : 'Wireframe' }}
      </button>
      <button 
        @click="resetCamera" 
        class="px-3 py-2 bg-white bg-opacity-90 rounded shadow hover:bg-opacity-100 text-sm"
      >
        Reset View
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as THREE from 'three'
import { STLLoader } from 'three-stdlib'
import { OrbitControls } from 'three-stdlib'

const props = defineProps({
  fileUrl: {
    type: String,
    required: true,
  },
  color: {
    type: Number,
    default: 0x0088ff,
  },
  showChanges: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['cameraChange'])

const container = ref(null)
const loading = ref(true)
const wireframe = ref(false)

let scene, camera, renderer, controls, mesh, initialCameraPosition

onMounted(() => {
  initScene()
  loadModel()
  animate()
  window.addEventListener('resize', onWindowResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onWindowResize)
  if (renderer) {
    renderer.dispose()
  }
})

watch(() => props.fileUrl, () => {
  loadModel()
})

const initScene = () => {
  // Scene
  scene = new THREE.Scene()
  scene.background = new THREE.Color(0x1a1a1a)

  // Camera
  camera = new THREE.PerspectiveCamera(
    45,
    container.value.clientWidth / container.value.clientHeight,
    0.1,
    10000
  )
  camera.position.set(200, 200, 200)
  initialCameraPosition = camera.position.clone()

  // Renderer
  renderer = new THREE.WebGLRenderer({ antialias: true })
  renderer.setSize(container.value.clientWidth, container.value.clientHeight)
  renderer.setPixelRatio(window.devicePixelRatio)
  container.value.appendChild(renderer.domElement)

  // Lights
  const ambientLight = new THREE.AmbientLight(0xffffff, 0.5)
  scene.add(ambientLight)

  const directionalLight1 = new THREE.DirectionalLight(0xffffff, 0.8)
  directionalLight1.position.set(1, 1, 1)
  scene.add(directionalLight1)

  const directionalLight2 = new THREE.DirectionalLight(0xffffff, 0.5)
  directionalLight2.position.set(-1, -1, -1)
  scene.add(directionalLight2)

  // Controls
  controls = new OrbitControls(camera, renderer.domElement)
  controls.enableDamping = true
  controls.dampingFactor = 0.05
  controls.addEventListener('change', () => {
    emit('cameraChange', {
      position: camera.position.clone(),
      target: controls.target.clone(),
    })
  })

  // Grid
  const gridHelper = new THREE.GridHelper(200, 20, 0x444444, 0x222222)
  scene.add(gridHelper)

  // Axes
  const axesHelper = new THREE.AxesHelper(100)
  scene.add(axesHelper)
}

const loadModel = async () => {
  loading.value = true

  // Remove existing mesh
  if (mesh) {
    scene.remove(mesh)
    mesh.geometry.dispose()
    mesh.material.dispose()
  }

  const loader = new STLLoader()
  
  try {
    const geometry = await new Promise((resolve, reject) => {
      loader.load(
        props.fileUrl,
        (geo) => resolve(geo),
        undefined,
        (error) => reject(error)
      )
    })

    // Center geometry
    geometry.computeBoundingBox()
    const center = new THREE.Vector3()
    geometry.boundingBox.getCenter(center)
    geometry.translate(-center.x, -center.y, -center.z)

    // Create material
    const material = new THREE.MeshPhongMaterial({
      color: props.color,
      specular: 0x111111,
      shininess: 200,
      wireframe: wireframe.value,
    })

    // Create mesh
    mesh = new THREE.Mesh(geometry, material)
    scene.add(mesh)

    // Fit camera to model
    fitCameraToModel(geometry)

    loading.value = false
  } catch (error) {
    console.error('Failed to load STL:', error)
    loading.value = false
  }
}

const fitCameraToModel = (geometry) => {
  const box = new THREE.Box3().setFromBufferAttribute(geometry.attributes.position)
  const size = box.getSize(new THREE.Vector3())
  const maxDim = Math.max(size.x, size.y, size.z)
  const fov = camera.fov * (Math.PI / 180)
  let cameraZ = Math.abs(maxDim / 2 / Math.tan(fov / 2))
  cameraZ *= 1.5 // Zoom out a bit

  camera.position.set(cameraZ, cameraZ, cameraZ)
  camera.lookAt(0, 0, 0)
  controls.target.set(0, 0, 0)
  controls.update()
}

const toggleWireframe = () => {
  wireframe.value = !wireframe.value
  if (mesh) {
    mesh.material.wireframe = wireframe.value
  }
}

const resetCamera = () => {
  camera.position.copy(initialCameraPosition)
  controls.target.set(0, 0, 0)
  controls.update()
}

const animate = () => {
  requestAnimationFrame(animate)
  controls.update()
  renderer.render(scene, camera)
}

const onWindowResize = () => {
  if (!container.value) return
  camera.aspect = container.value.clientWidth / container.value.clientHeight
  camera.updateProjectionMatrix()
  renderer.setSize(container.value.clientWidth, container.value.clientHeight)
}

// Expose methods for parent component
defineExpose({
  syncCamera: (position, target) => {
    camera.position.copy(position)
    controls.target.copy(target)
    controls.update()
  },
})
</script>