<script setup lang="ts">
import { Fieldset as ArkFieldset } from '@ark-ui/vue/fieldset'
import { computed, provide } from 'vue'
import { switchGroupKey } from './context'

const props = withDefaults(defineProps<{ legend?: string, description?: string, error?: string, disabled?: boolean, controlFirst?: boolean }>(), { controlFirst: true })
provide(switchGroupKey, { disabled: computed(() => props.disabled ?? false), controlFirst: computed(() => props.controlFirst) })
</script>

<template>
  <ArkFieldset.Root :disabled="props.disabled" :invalid="!!props.error" class="flex flex-col gap-3 rounded-lg">
    <ArkFieldset.Legend v-if="props.legend" class="mb-3 text-size-base font-medium text-default">
      {{ props.legend }}
    </ArkFieldset.Legend>
    <slot />
    <ArkFieldset.ErrorText v-if="props.error" class="text-size-sm text-fg-danger">
      {{ props.error }}
    </ArkFieldset.ErrorText>
    <ArkFieldset.HelperText v-if="props.description" class="text-size-sm text-subtle">
      {{ props.description }}
    </ArkFieldset.HelperText>
  </ArkFieldset.Root>
</template>
