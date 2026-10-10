<script setup lang="ts">
import type { BundledLanguage } from 'shiki'
import { onBeforeUnmount, shallowRef, watch } from 'vue'

const props = withDefaults(defineProps<{ code: string, lang?: BundledLanguage | 'text', lineNumbers?: boolean }>(), { lang: 'text' })
const html = shallowRef('')
let revision = 0
watch(() => [props.code, props.lang] as const, async ([code, lang]) => {
  const current = ++revision
  html.value = ''
  try {
    const { codeToHtml } = await import('shiki')
    const rendered = await codeToHtml(code, { lang, themes: { light: 'github-light', dark: 'github-dark' }, defaultColor: false })
    if (current === revision)
      html.value = rendered
  }
  catch {
    // Keep readable source code if loading a grammar fails.
    if (current === revision)
      html.value = ''
  }
}, { immediate: true })
onBeforeUnmount(() => {
  revision++
})
</script>

<template>
  <!-- Shiki escapes source code before producing this trusted markup. -->
  <div v-if="html" class="kumo-code overflow-auto rounded-lg bg-recessed p-3 font-mono text-size-sm [&_pre]:m-0 [&_pre]:bg-transparent!" :class="lineNumbers && 'with-line-numbers'" v-html="html" />
  <pre v-else class="overflow-auto rounded-lg bg-recessed p-3 font-mono text-size-sm"><code>{{ code }}</code></pre>
</template>

<style>
.kumo-code .shiki,
.kumo-code .shiki span {
  color: var(--shiki-light);
}
[data-mode='dark'] .kumo-code .shiki,
[data-mode='dark'] .kumo-code .shiki span {
  color: var(--shiki-dark);
}
.kumo-code.with-line-numbers code {
  counter-reset: code-line;
}
.kumo-code.with-line-numbers .line::before {
  display: inline-block;
  min-width: 2em;
  margin-right: 1em;
  color: var(--text-color-subtle);
  text-align: right;
  counter-increment: code-line;
  content: counter(code-line);
}
</style>
