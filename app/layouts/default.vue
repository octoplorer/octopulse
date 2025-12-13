<template>
  <div class="min-h-screen flex flex-col">
    <header
      class="bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 sticky top-0 z-30 transition-colors duration-300"
    >
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="bg-indigo-600 p-2 rounded-lg">
            <i class="i-lucide:shield-check text-white w-6 h-6" aria-hidden="true" />
          </div>
          <h1 class="text-xl font-bold tracking-tight text-slate-900 dark:text-white">
            NuxtPulse
          </h1>
        </div>
        <div class="flex items-center gap-4">
          <div class="hidden md:flex flex-col items-end mr-2">
            <span class="text-xs text-slate-500 dark:text-slate-400 uppercase font-semibold tracking-wider">最后更新</span>
            <span class="text-sm font-medium tabular-nums text-slate-900 dark:text-slate-200">03:17:06</span>
          </div>
          <button
            class="p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500 dark:text-slate-400 transition-all"
            title="Switch to Dark Mode"
          >
            <div class="i-lucide:moon w-5 h-5" aria-hidden="true" />
          </button>
          <button
            class="p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 transition-all text-slate-500 dark:text-slate-400"
            title="Refresh Status"
          >
            <div class="i-lucide:refresh-cw w-5 h-5" aria-hidden="true" />
          </button>
        </div>
      </div>
    </header>

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
