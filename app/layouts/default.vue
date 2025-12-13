<script setup lang="ts">
</script>

<template>
  <div
    class="min-h-screen bg-zinc-50 dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100 pb-12 transition-colors duration-300"
  >
    <HomeHeaders />

    <main class="flex-grow max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <slot />
    </main>

    <footer class="flex-shrink-0 max-w-7xl mx-auto px-4 py-6 text-center text-slate-400 dark:text-slate-500 text-sm transition-colors">
      <div class="text-center text-gray-400 text-sm">
        © 2025 NuxtPulse. All systems operational.
      </div>
    </footer>
  </div>
</template>
