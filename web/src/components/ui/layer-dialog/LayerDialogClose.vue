<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
import { Button } from '../button'
import { useLayerDialog } from './context'

const context = useLayerDialog()
</script>

<template>
  <ArkDialog.CloseTrigger as-child>
    <Button :disabled="context?.dismissDisabled.value">
      <slot>Cancel</slot>
    </Button>
  </ArkDialog.CloseTrigger>
</template>
