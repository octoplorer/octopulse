import type { User } from './client/types.gen'
import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    roles?: User['role'][]
  }
}
