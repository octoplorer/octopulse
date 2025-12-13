<script setup lang="ts">
const { data, isLoading } = useQuery({
  key: ['overview'],
  query: () => $fetch('/api/services/overviews'),
})
const queryCache = useQueryCache()
const colorMode = useColorMode()
const { loggedIn, clear } = useUserSession()

function onThemeToggle() {
  colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'
}
</script>

<template>
  <header
    bg="white dark:zinc-900"
    border="b zinc-200 dark:zinc-800"
    class="sticky top-0 z-30 transition-colors duration-300"
  >
    <div
      flex="~ items-center justify-end"
      max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16
    >
      <div flex="~ items-center gap-4">
        <div flex="md:~ col items-end" mr-2hidden>
          <span
            un-text="xs zinc-500 dark:zinc-400"
            uppercase font-semibold tracking-wider
          >
            最后更新
          </span>
          <span
            un-text="zinc-900 dark:zinc-200"
            text-sm font-medium tabular-nums
          >
            <NuxtTime
              v-if="data" :datetime="data.timestamp"
              day="numeric" month="long"
              hour="2-digit" minute="2-digit" second="2-digit"
            />
          </span>
        </div>
        <button
          bg="hover:zinc-100 dark:hover:zinc-800" p-2 rounded-full
          un-text="zinc-500 dark:zinc-400"
          class="transition-all"
          title="Toggle color theme"
          @click="onThemeToggle"
        >
          <ClientOnly>
            <div :class="$colorMode.value === 'dark' ? 'i-lucide:moon' : 'i-lucide:sun'" size-5 aria-hidden="true" />
            <template #fallback>
              <div size-5 i-lucide:earth aria-hidden="true" />
            </template>
          </ClientOnly>
        </button>

        <button
          v-if="$route.path !== '/admin'"
          :data-sign-in="$route.path === '/sign-in'"
          bg="hover:zinc-100 dark:hover:zinc-800" p-2 rounded-full
          un-text="zinc-500 dark:zinc-400 data-[sign-in]:text-indigo-600"
          class="transition-all"
          title="Admin login"
          @click="() => {
            $router.push('/admin')
          }"
        >
          <div i-lucide:settings size-5 aria-hidden="true" />
        </button>

        <button
          v-if="loggedIn"
          bg="hover:zinc-100 dark:hover:zinc-800" p-2 rounded-full
          un-text="zinc-500 dark:zinc-400"
          class="transition-all"
          title="Admin login"
          @click="() => {
            clear().then(() => {
              $router.push('/')
            })
          }"
        >
          <div i-lucide:log-out size-5 aria-hidden="true" />
        </button>

        <button
          v-if="$route.path === '/'"
          :data-load="isLoading"
          bg="hover:zinc-100 dark:hover:zinc-800" p-2 rounded-full
          un-text="zinc-500 dark:zinc-400 data-[loading]:text-indigo-600"
          class="transition-all"
          :class="isLoading ? 'animate-spin' : ''"
          title="Refresh Status"
          :disabled="isLoading"
          @click="() => {
            queryCache.invalidateQueries({ key: ['overview'] })
          }"
        >
          <div i-lucide:refresh-cw size-5 aria-hidden="true" />
        </button>
        <button
          v-if="loggedIn"
          class="ml-2 px-3 py-1.5 text-xs font-medium bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg transition-colors"
        >
          管理服务
        </button>
      </div>
    </div>
  </header>
</template>
