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

    <main flex-grow max-w-7xl mx-auto p="x4 sm:x6 lg:x8 y8">
      <slot />
    </main>

    <footer
      un-text="center zinc-500 dark:zinc-300 sm"
      flex-shrink-0 max-w-7xl mx-auto px-4 py-6 transition-colors
    >
      <div text-center text-sm>
        © 2025 NuxtPulse. All systems operational.
      </div>
    </footer>
  </div>
</template>
