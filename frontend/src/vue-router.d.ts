export {}

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    icon?: unknown
    requiresAuth?: boolean
    requiresControlPanel?: boolean
    requiresIam?: boolean
    requiresSettings?: boolean
    breadcrumbs?: (route: unknown) => Array<{ name: string; href?: string; to?: { name: string } }>
  }
}
