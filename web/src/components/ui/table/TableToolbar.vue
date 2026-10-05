<template>
  <div flex="~ justify-between items-center gap-12px" p="y-16px x-20px" border="b-1 solid $border" class="[@media(max-width:700px)]:flex-wrap [@media(max-width:700px)]:px-15px [@media(max-width:700px)]:py-13px [@media(max-width:700px)]:[&>div:not(.search-box)]:max-w-full [@media(max-width:700px)]:[&>div:not(.search-box)]:flex-wrap [@container_workspace_(max-width:_700px)]:flex-wrap">
    <slot />
  </div>
</template>
