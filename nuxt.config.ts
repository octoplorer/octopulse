// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: [
    '@nuxthub/core',
    '@unocss/nuxt',
    '@vueuse/nuxt',
    '@nuxt/eslint',
    '@nuxtjs/color-mode',
    '@pinia/nuxt',
    '@pinia/colada-nuxt',
  ],

  components: {
    global: true,
    dirs: [
      {
        path: '~/components/modules',
      },
      '~/components',
    ],
  },

  hub: {
    db: 'sqlite',
    kv: true,
  },

  eslint: {
    config: {
      standalone: false,
    },
  },

  nitro: {
    experimental: {
      tasks: true,
    },
    scheduledTasks: {
      '* * * * *': [
        'check',
        'incident',
      ],
    },
    typescript: {
      tsConfig: {
        compilerOptions: {
          types: ['temporal-polyfill/global'],
        },
      },
    },
    moduleSideEffects: ['temporal-polyfill/global'],
    minify: false,
  },
  typescript: {
    sharedTsConfig: {
      compilerOptions: {
        types: ['temporal-polyfill/global'],
      },
    },
    tsConfig: {
      compilerOptions: {
        types: ['temporal-polyfill/global'],
      },
    },
  },
})
