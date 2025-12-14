<script setup lang="ts">
useHead({
  bodyAttrs: {
    class: 'bg-zinc-50 dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100 transition-colors duration-300',
  },
})
</script>

<template>
  <div
    flex="~ col"
    class="min-h-screen"
  >
    <HomeHeaders />

    <main flex-grow max-w-80rem w-full mx-auto p="x4 sm:x6 lg:x8 y8">
      <slot />
    </main>

    <footer
      w-full
      un-text="center zinc-500 dark:zinc-300 sm"
      flex-shrink-0 max-w-7xl mx-auto px-4 py-6 transition-colors
      class="mt-16 pt-6 border-t border-zinc-200 dark:border-zinc-800"
    >
      <div
        flex="~ col sm:row items-center justify-between gap-4"
        un-text="xs zinc-500 dark:zinc-400"
      >
        <div flex="~ items-center gap-4">
          <span font-mono>Powered By Octopulse</span>
        </div>
      </div>
    </footer>
  </div>
</template>
