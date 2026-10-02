<script setup lang="ts">
import MonitorEditor from '../../../../../components/MonitorEditor.vue'

definePage({ meta: { title: ['编辑监控项', 'Edit monitor'], roles: ['admin', 'operator'] } })
</script>

<template>
  <MonitorEditor />
</template>
