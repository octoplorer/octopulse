<script setup lang="ts">
import MonitorEditor from '../../../../components/MonitorEditor.vue'

definePage({ meta: { title: ['创建监控项', 'New monitor'], roles: ['admin', 'operator'] } })
</script>

<template>
  <MonitorEditor />
</template>
