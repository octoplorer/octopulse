<script setup lang="ts">
const password = ref('')
const { fetch } = useUserSession()
const { error, isLoading, mutateAsync } = useMutation({
  mutation(vars: string) {
    return $fetch('/api/login', {
      method: 'POST',
      body: {
        password: vars,
      },
    })
  },
  async onSuccess() {
    await fetch()
    await navigateTo('/admin')
  },
})

function handleReturnToDashboard() {
  navigateTo('/')
}
</script>

<template>
  <div
    flex="~ items-center justify-center"
    class="min-h-[calc(100vh-12rem)]"
  >
    <div
      bg="white dark:zinc-900"
      border="~ zinc-200 dark:zinc-800 rounded-2xl"
      class="w-full max-w-md mx-4 p-8 shadow-xl relative z-10 transition-all"
    >
      <!-- Icon and Title Section -->
      <div text-center mb-8>
        <div
          mx-auto
          h-14
          w-14
          bg="indigo-50 dark:indigo-900/20"
          un-text="indigo-600 dark:indigo-400"
          rounded-2xl
          flex="~ items-center justify-center"
          mb-5
          ring="1 indigo-100 dark:indigo-900/30"
        >
          <div size-7.5 i-lucide:shield-check />
        </div>
        <h2
          un-text="2xl zinc-900 dark:zinc-100"
          font-bold
          tracking-tight
        >
          管理员验证
        </h2>
        <p
          mt-2
          un-text="sm zinc-500 dark:zinc-400"
        >
          请输入访问密码以管理监控服务
        </p>
      </div>

      <!-- Form -->
      <form
        space-y-6
        @submit.prevent="() => {
          mutateAsync(password)
        }"
      >
        <!-- Password Input -->
        <div>
          <label for="password" sr-only>密码</label>
          <div relative group>
            <div
              absolute inset-y-0 left-0 pl-3
              flex="~ items-center"
              pointer-events-none transition-colors
              un-text="zinc-400 group-focus-within:indigo-500"
            >
              <div size-4.5 i-lucide:lock />
            </div>
            <input
              id="password"
              v-model="password"
              required
              type="password"
              name="password"
              placeholder="输入管理员密码"
              appearance-none block w-full
              p="l10 r3 y3"
              border="~ zinc-200 dark:zinc-700 focus:indigo-500 rounded-xl"
              ring="focus:indigo-500 focus:2 focus:offset-1  dark:focus:offset-zinc-900"
              bg="zinc-50 dark:zinc-950"
              un-text="sm:sm placeholder:zinc-400 zinc-900 dark:white"
              focus:outline-none
              transition-all
            >
          </div>
          <p
            v-if="error"
            un-text="sm rose-500"
            mt-2
          >
            {{ error }}
          </p>
        </div>

        <!-- Buttons -->
        <div flex="~ col gap-3" pt-2>
          <button
            type="submit"
            w-full flex="~ justify-center"
            p="y2.5 x4"
            border="~ transparent rounded-xl"
            un-text="sm white"
            font-semibold
            bg="indigo-600 hover:indigo-700"
            focus:outline-none
            ring="focus:2 focus:offset-2 focus:indigo-500"
            class="transition-all shadow-lg shadow-indigo-500/20 hover:shadow-indigo-500/30 hover:-translate-y-0.5 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0"
            :disabled="isLoading"
          >
            {{ isLoading ? '验证中...' : '解锁控制台' }}
          </button>
          <button
            type="button"
            w-full flex="~ justify-center"
            p="y2.5 x4" font-medium
            un-text="sm zinc-500 hover:zinc-800 dark:zinc-400 dark:hover:zinc-200"
            transition-colors
            @click="handleReturnToDashboard"
          >
            返回仪表盘
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
