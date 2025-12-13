<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})

// Form state
const serviceName = ref('')
const serviceUrl = ref('')
const serviceTags = ref('')

// Fetch services
const { data: servicesData, isLoading, refetch } = useQuery({
  key: ['admin-services'],
  query() {
    // Don't known why it need
    return $fetch('/api/services', { headers: useRequestHeaders() })
  },
})

// Create service mutation
const { mutateAsync: createService, isLoading: isCreating } = useMutation({
  mutation(vars: { name: string, url: string, tags: string[] }) {
    return $fetch('/api/services', {
      method: 'POST',
      body: {
        name: vars.name,
        url: vars.url,
        tags: vars.tags,
        method: 'GET',
        type: 'public',
      },
    })
  },
  async onSuccess() {
    // Reset form
    serviceName.value = ''
    serviceUrl.value = ''
    serviceTags.value = ''
    // Refresh services list
    await refetch()
  },
})

// Delete service mutation
const { mutateAsync: deleteService, isLoading: isDeleting } = useMutation({
  mutation(vars: number) {
    return $fetch(`/api/services/${vars}`, {
      method: 'DELETE',
    })
  },
  async onSuccess() {
    // Refresh services list
    await refetch()
  },
})

// Handle form submission
async function handleSubmit() {
  if (!serviceName.value || !serviceUrl.value) {
    return
  }

  const tags = serviceTags.value
    .split(',')
    .map(tag => tag.trim())
    .filter(tag => tag.length > 0)

  try {
    await createService({
      name: serviceName.value,
      url: serviceUrl.value,
      tags,
    })
  }
  catch (error) {
    console.error('Failed to create service:', error)
  }
}

// Handle service deletion
async function handleDelete(serviceId: number) {
  try {
    await deleteService(serviceId)
  }
  catch (error) {
    console.error('Failed to delete service:', error)
  }
}

// Navigate back
function handleBack() {
  navigateTo('/')
}

const services = computed(() => servicesData.value?.services ?? [])
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 py-8">
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <div class="flex items-center gap-4">
        <button
          class="p-2 -ml-2 rounded-full hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-500 transition-colors"
          @click="handleBack"
        >
          <div size-6 i-lucide:arrow-left aria-hidden="true" />
        </button>
        <div>
          <h1 class="text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            服务管理
          </h1>
          <p class="text-zinc-500 dark:text-zinc-400 text-sm">
            添加或移除监控端点
          </p>
        </div>
      </div>
    </div>

    <!-- Main Content Grid -->
    <div class="grid gap-8 md:grid-cols-2">
      <!-- Add New Service Form -->
      <div class="bg-white dark:bg-zinc-900 rounded-xl shadow-sm border border-zinc-200 dark:border-zinc-800 p-6 h-fit">
        <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100 mb-4 flex items-center gap-2">
          <div size-5 i-lucide:plus class="text-indigo-500" aria-hidden="true" />
          添加新服务
        </h2>
        <form class="space-y-4" @submit.prevent="handleSubmit">
          <!-- Service Name -->
          <div>
            <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              服务名称
            </label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <div size-4 i-lucide:server class="text-zinc-400" aria-hidden="true" />
              </div>
              <input
                v-model="serviceName"
                required
                type="text"
                placeholder="例如: Payment API"
                class="block w-full pl-10 pr-3 py-2 bg-zinc-50 dark:bg-zinc-950 border border-zinc-300 dark:border-zinc-700 rounded-lg focus:ring-indigo-500 focus:border-indigo-500 dark:text-white sm:text-sm outline-none transition-colors"
              >
            </div>
          </div>

          <!-- Monitoring URL -->
          <div>
            <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              监控 URL
            </label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <div size-4 i-lucide:globe class="text-zinc-400" aria-hidden="true" />
              </div>
              <input
                v-model="serviceUrl"
                required
                type="text"
                placeholder="例如: https://api.example.com"
                class="block w-full pl-10 pr-3 py-2 bg-zinc-50 dark:bg-zinc-950 border border-zinc-300 dark:border-zinc-700 rounded-lg focus:ring-indigo-500 focus:border-indigo-500 dark:text-white sm:text-sm outline-none transition-colors"
              >
            </div>
          </div>

          <!-- Tags -->
          <div>
            <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              标签
              <span class="text-zinc-400 font-normal">(可选，逗号分隔)</span>
            </label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <div size-4 i-lucide:tag class="text-zinc-400" aria-hidden="true" />
              </div>
              <input
                v-model="serviceTags"
                type="text"
                placeholder="例如: Production, API, Core"
                class="block w-full pl-10 pr-3 py-2 bg-zinc-50 dark:bg-zinc-950 border border-zinc-300 dark:border-zinc-700 rounded-lg focus:ring-indigo-500 focus:border-indigo-500 dark:text-white sm:text-sm outline-none transition-colors"
              >
            </div>
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="isCreating"
            class="w-full flex justify-center py-2 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ isCreating ? '添加中...' : '添加监控' }}
          </button>
        </form>
      </div>

      <!-- Monitored Services List -->
      <div class="bg-white dark:bg-zinc-900 rounded-xl shadow-sm border border-zinc-200 dark:border-zinc-800 p-6">
        <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100 mb-4">
          已监控服务 ({{ services.length }})
        </h2>
        <div class="space-y-3 max-h-[400px] overflow-y-auto pr-2">
          <div
            v-if="isLoading"
            class="flex items-center justify-center py-8"
          >
            <span class="text-zinc-500 dark:text-zinc-400 text-sm">加载中...</span>
          </div>
          <div
            v-else-if="services.length === 0"
            class="flex items-center justify-center py-8"
          >
            <span class="text-zinc-500 dark:text-zinc-400 text-sm">暂无服务</span>
          </div>
          <div
            v-for="service in services"
            :key="service.id"
            class="flex items-center justify-between p-3 rounded-lg bg-zinc-50 dark:bg-zinc-950 border border-zinc-100 dark:border-zinc-800 group"
          >
            <div class="min-w-0 flex-1">
              <h3 class="text-sm font-medium text-zinc-900 dark:text-zinc-200 truncate">
                {{ service.name }}
              </h3>
              <p class="text-xs text-zinc-500 dark:text-zinc-400 truncate mb-1">
                {{ service.url }}
              </p>
            </div>
            <button
              :disabled="isDeleting"
              class="p-2 text-zinc-400 hover:text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-900/20 rounded-lg transition-colors opacity-0 group-hover:opacity-100 focus:opacity-100 disabled:opacity-50 disabled:cursor-not-allowed"
              title="删除服务"
              @click="handleDelete(service.id)"
            >
              <div size-4 i-lucide:trash-2 aria-hidden="true" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
