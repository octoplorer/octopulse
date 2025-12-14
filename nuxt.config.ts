import process from 'node:process'

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
    'nuxt-auth-utils',
    'reka-ui',
  ],

  runtimeConfig: {
    session: {
      name: 'nuxt-pulse-session',
      password: process.env.NUXT_SESSION_PASSWORD || '',
      cookie: {
        sameSite: 'lax',
      },
      maxAge: 60 * 60 * 24 * 7, // 1 week
    },
  },

  components: {
    global: true,
    dirs: [
      {
        path: '~/components/modules',
      },
      '~/components',
    ],
  },

  imports: {
    dirs: ['queries'],
  },

  hub: {
    db: 'sqlite',
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
      ],
    },
    // https://github.com/nitrojs/nitro/issues/1974
    // workaround
    vercel: {
      config: {
        crons: [
          {
            path: '/api/_cron',
            schedule: '* * * * *',
          },
        ],
      },
    },
    typescript: {
      tsConfig: {
        compilerOptions: {
          types: ['temporal-polyfill/global'],
        },
      },
    },
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
