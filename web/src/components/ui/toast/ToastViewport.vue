<template>
  <div position="fixed" bottom="24px" right="24px" z="500" flex="~ col gap-9px" max-w="430px" aria-live="polite" class="[@media(max-width:700px)]:bottom-16px [@media(max-width:700px)]:left-16px [@media(max-width:700px)]:right-16px [@media(max-width:700px)]:max-w-none">
    <slot />
  </div>
</template>
