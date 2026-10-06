<script setup lang="ts">
import type { UseTagsInputProps } from '@ark-ui/vue/tags-input'
import { useFieldContext } from '@ark-ui/vue/field'
import { TagsInput as ArkTagsInput, useTagsInput } from '@ark-ui/vue/tags-input'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { splitValues } from '../lib/form'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  name?: string
  placeholder?: string
  disabled?: boolean
  readOnly?: boolean
  required?: boolean
  form?: string
}>(), { disabled: undefined, readOnly: undefined, required: undefined })
const value = defineModel<string[]>({ required: true })
const { t } = useI18n({ useScope: 'global' })
const field = useFieldContext()
const tagsInput = useTagsInput(computed<UseTagsInputProps>(() => ({
  ...props,
  modelValue: value.value,
  ids: { input: field?.value.ids.control, label: field?.value.ids.label },
  addOnPaste: true,
  allowDuplicates: true,
  blurBehavior: 'add',
  delimiter: /[,\n]/,
  translations: {
    deleteTagTriggerLabel: (tag: string) => `${t('common.delete')} ${tag}`,
  },
  onValueChange: ({ value: tags }) => value.value = tags,
})))

function commit() {
  const pending = splitValues(tagsInput.value.inputValue)
  if (pending.length)
    tagsInput.value.setValue([...tagsInput.value.value, ...pending])
  tagsInput.value.clearInputValue()
}
defineExpose({ commit })
</script>

<template>
  <ArkTagsInput.RootProvider v-bind="$attrs" :value="tagsInput">
    <ArkTagsInput.Control class="tags-control">
      <ArkTagsInput.Item v-for="(tag, index) in value" :key="index" :index="index" :value="tag" class="tags-item">
        <ArkTagsInput.ItemPreview flex="~ items-center gap-1">
          <ArkTagsInput.ItemText>{{ tag }}</ArkTagsInput.ItemText>
          <ArkTagsInput.ItemDeleteTrigger bg="transparent" border="0" p="0" flex="~ items-center" un-text="subtle">
            <span class="i-lucide-x" size="12px" aria-hidden="true" />
          </ArkTagsInput.ItemDeleteTrigger>
        </ArkTagsInput.ItemPreview>
        <ArkTagsInput.ItemInput :aria-label="`${t('common.edit')} ${tag}`" />
      </ArkTagsInput.Item>
      <ArkTagsInput.Input />
    </ArkTagsInput.Control>
    <ArkTagsInput.HiddenInput />
  </ArkTagsInput.RootProvider>
</template>

<style scoped>
.tags-control {
  display: flex;
  min-height: 36px;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  background: var(--color-control);
  border: 1px solid var(--color-line);
  border-radius: 8px;
  box-shadow: var(--shadow-control);
}

.tags-control:focus-within {
  outline: 2px solid var(--color-focus);
  outline-offset: 3px;
}

.tags-item {
  padding: 2px 6px;
  color: var(--text-color-default);
  font-size: 13px;
  background: var(--color-recessed);
  border-radius: 5px;
}

.tags-control :deep(input[data-scope='tags-input'][data-part]) {
  width: auto;
  min-width: 80px;
  min-height: 24px;
  flex: 1;
  padding: 0;
  background: transparent;
  border: 0;
  border-radius: 0;
  outline: none;
  box-shadow: none;
}

.tags-control[data-disabled] {
  opacity: 0.5;
}
</style>
