<script setup lang="ts">
import { useElementBounding, useResizeObserver, useWindowSize } from '@vueuse/core'
import { nextTick, shallowRef, useId } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from './ui/button'

const { t } = useI18n({ useScope: 'global' })
const activePane = shallowRef<'edit' | 'preview'>('edit')
const id = useId()
const editorPane = shallowRef<HTMLElement>()
const workspaceElement = shallowRef<HTMLElement>()
const previewPane = shallowRef<HTMLElement>()
const { top: previewTop, update: updatePreviewBounds } = useElementBounding(previewPane, { updateTiming: 'next-frame' })
const { height: viewportHeight } = useWindowSize()
useResizeObserver(() => workspaceElement.value?.parentElement, updatePreviewBounds)
async function focusEditor(name: string) {
  activePane.value = 'edit'
  await nextTick()
  editorPane.value?.querySelector<HTMLElement>(`[name="${name}"]`)?.focus()
}
defineExpose({ focusEditor })
</script>

<template>
  <div ref="workspaceElement" class="page-editor-workspace">
    <div class="editor-view-switch mb-5 flex w-fit gap-1 rounded-lg bg-recessed p-1" role="group" :aria-label="t('page-editor.editor-views')">
      <Button :variant="activePane === 'edit' ? 'secondary' : 'ghost'" :aria-pressed="activePane === 'edit'" :aria-controls="`${id}-edit`" @click="activePane = 'edit'">
        <span class="i-lucide-pencil size-4" aria-hidden="true" />{{ t('common.edit') }}
      </Button>
      <Button :variant="activePane === 'preview' ? 'secondary' : 'ghost'" :aria-pressed="activePane === 'preview'" :aria-controls="`${id}-preview`" @click="activePane = 'preview'">
        <span class="i-lucide-eye size-4" aria-hidden="true" />{{ t('page-editor.preview') }}
      </Button>
    </div>
    <div class="editor-workspace grid min-w-0 items-start gap-6">
      <div v-show="activePane === 'edit'" :id="`${id}-edit`" ref="editorPane" class="editor-pane min-w-0">
        <slot name="editor" />
      </div>
      <aside
        v-show="activePane === 'preview'" :id="`${id}-preview`" ref="previewPane" class="preview-pane min-w-0"
        :style="{ '--preview-top': `${previewTop}px`, '--preview-viewport-height': `${viewportHeight}px` }"
        :aria-label="t('page-editor.live-draft-preview')" tabindex="0"
      >
        <slot name="preview" />
      </aside>
    </div>
  </div>
</template>

<style scoped>
.page-editor-workspace {
  container: page-editor-workspace / inline-size;
}

.editor-pane {
  container: status-page-editor / inline-size;
}

@container page-editor-workspace (min-width: 1000px) {
  .editor-view-switch {
    display: none;
  }

  .editor-workspace {
    grid-template-columns: minmax(0, 3fr) minmax(0, 2fr);
  }

  .editor-pane,
  .preview-pane {
    /* Wide layouts show both mounted panes, regardless of the narrow view selection. */
    display: block !important;
  }

  .preview-pane {
    position: sticky;
    top: calc(var(--workspace-header-height) + 16px);
    max-height: max(0px, calc(var(--preview-viewport-height) - var(--preview-top) - 6rem));
    overflow: auto;
    overscroll-behavior: contain;
  }
}
</style>
