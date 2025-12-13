<script setup lang="ts">
const queryCache = useQueryCache()
</script>

<template>
  <header
    class="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 sticky top-0 z-30 transition-colors duration-300"
  >
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-end">
      <div class="flex items-center gap-4">
        <div class="hidden md:flex flex-col items-end mr-2">
          <span class="text-xs text-zinc-500 dark:text-zinc-400 uppercase font-semibold tracking-wider">
            最后更新
          </span>
          <span
            class="text-sm font-medium tabular-nums text-zinc-900 dark:text-zinc-200"
          >
            16:32:03
          </span>
        </div>
        <button
          class="p-2 rounded-full hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-500 dark:text-zinc-400 transition-all"
          :title="`Switch to ${$colorMode.value === 'dark' ? 'Light' : 'Dark'} Mode`"
          @click="() => $colorMode.preference = $colorMode.value === 'dark' ? 'light' : 'dark'"
        >
          <ClientOnly>
            <div :class="$colorMode.value === 'dark' ? 'i-lucide:moon' : 'i-lucide:sun'" size-5 aria-hidden="true" />
            <template #fallback>
              <div size-5 i-lucide:eclipse aria-hidden="true" />
            </template>
          </ClientOnly>
        </button>
        <button
          class="p-2 rounded-full hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-all text-zinc-500 dark:text-zinc-400"
          title="Refresh Status"
          @click="() => {
            queryCache.invalidateQueries({ key: ['overview'] })
          }"
        >
          <div class="i-lucide:refresh-cw w-5 h-5" aria-hidden="true" />
        </button>
      </div>
    </div>
  </header>
</template>
