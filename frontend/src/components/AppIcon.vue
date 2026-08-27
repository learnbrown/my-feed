<script setup>
defineProps({
  name: {
    type: String,
    required: true
  },
  size: {
    type: [Number, String],
    default: 20
  }
});
</script>

<template>
  <svg
    class="app-icon"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.9"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <template v-if="name === 'compass'">
      <circle cx="12" cy="12" r="9" />
      <path d="m15.5 8.5-2.1 4.9-4.9 2.1 2.1-4.9 4.9-2.1Z" />
    </template>
    <template v-else-if="name === 'upload'">
      <path d="M12 16V4" />
      <path d="m7.5 8.5 4.5-4.5 4.5 4.5" />
      <path d="M5 14v5h14v-5" />
    </template>
    <template v-else-if="name === 'user'">
      <circle cx="12" cy="8" r="3.5" />
      <path d="M5.5 20c.5-4 2.7-6 6.5-6s6 2 6.5 6" />
    </template>
    <template v-else-if="name === 'search'">
      <circle cx="10.5" cy="10.5" r="6.5" />
      <path d="m16 16 4 4" />
    </template>
    <template v-else-if="name === 'heart'">
      <path d="M20.8 5.8c-1.9-2-5-2-6.9 0L12 7.7l-1.9-1.9a4.8 4.8 0 0 0-6.9 6.7L12 21l8.8-8.5a4.8 4.8 0 0 0 0-6.7Z" />
    </template>
    <template v-else-if="name === 'comment'">
      <path d="M20 15a4 4 0 0 1-4 4H9l-5 3v-6a5 5 0 0 1-1-3V8a4 4 0 0 1 4-4h9a4 4 0 0 1 4 4v7Z" />
    </template>
    <template v-else-if="name === 'flame'">
      <path d="M13.5 3.5c.4 3-1.7 4.1-2.8 5.8-.9 1.3-.5 2.6.5 3.5-.2-2.4 1.8-3.4 2.9-4.6 1.3 2 3.4 3.8 3.4 6.7A5.5 5.5 0 0 1 12 20.5 5.5 5.5 0 0 1 6.5 15c0-4.2 3.2-6.2 7-11.5Z" />
    </template>
    <template v-else-if="name === 'info'">
      <circle cx="12" cy="12" r="9" />
      <path d="M12 11v5" />
      <path d="M12 8h.01" />
    </template>
    <template v-else-if="name === 'play'">
      <path d="m9 6 9 6-9 6V6Z" />
    </template>
    <template v-else-if="name === 'pause'">
      <path d="M9 6v12M15 6v12" />
    </template>
    <template v-else-if="name === 'volume'">
      <path d="M6 9H3v6h3l5 4V5L6 9Z" />
      <path d="M15 9a4 4 0 0 1 0 6M17.5 6.5a8 8 0 0 1 0 11" />
    </template>
    <template v-else-if="name === 'mute'">
      <path d="M6 9H3v6h3l5 4V5L6 9Z" />
      <path d="m16 10 5 5m0-5-5 5" />
    </template>
    <template v-else-if="name === 'mail'">
      <rect x="3" y="5" width="18" height="14" rx="3" />
      <path d="m4 7 8 6 8-6" />
    </template>
    <template v-else-if="name === 'user-plus'">
      <circle cx="9" cy="8" r="3" />
      <path d="M3.5 19c.4-3.6 2.2-5.4 5.5-5.4 1.7 0 3 .5 4 1.4M18 8v6M15 11h6" />
    </template>
    <template v-else-if="name === 'user-check'">
      <circle cx="9" cy="8" r="3" />
      <path d="M3.5 19c.4-3.6 2.2-5.4 5.5-5.4 1.7 0 3 .5 4 1.4m2-4 2 2 4-5" />
    </template>
    <template v-else-if="name === 'close'">
      <path d="M6 6l12 12M18 6 6 18" />
    </template>
    <template v-else-if="name === 'refresh'">
      <path d="M20 11a8 8 0 1 0-2.4 5.7" />
      <path d="M20 5v6h-6" />
    </template>
    <template v-else-if="name === 'logout'">
      <path d="M10 4H5v16h5M14 8l4 4-4 4M8 12h10" />
    </template>
    <template v-else-if="name === 'users'">
      <circle cx="9" cy="8" r="3" />
      <path d="M3.5 19c.4-3.7 2.2-5.5 5.5-5.5s5.1 1.8 5.5 5.5M15 5.5a3 3 0 0 1 0 5.5m1 3c2.7.3 4.2 2 4.5 5" />
    </template>
    <template v-else-if="name === 'film'">
      <rect x="3" y="4" width="18" height="16" rx="3" />
      <path d="M7 4v16M17 4v16M3 9h4m10 0h4M3 15h4m10 0h4" />
    </template>
    <template v-else-if="name === 'chevron-right'">
      <path d="m9 6 6 6-6 6" />
    </template>
    <template v-else-if="name === 'send'">
      <path d="m21 3-7.5 18-3-7.5L3 10.5 21 3Z" />
      <path d="m10.5 13.5 4-4" />
    </template>
    <template v-else-if="name === 'trash'">
      <path d="M4 7h16M9 7V4h6v3M7 7l1 13h8l1-13M10 11v5M14 11v5" />
    </template>
    <template v-else-if="name === 'image'">
      <rect x="3" y="4" width="18" height="16" rx="3" />
      <circle cx="9" cy="9" r="2" />
      <path d="m4 17 5-5 4 4 2-2 5 4" />
    </template>
    <template v-else-if="name === 'check'">
      <path d="m5 12 4 4L19 6" />
    </template>
    <template v-else-if="name === 'tag'">
      <path d="M20 13 13 20 4 11V4h7l9 9Z" />
      <circle cx="8.5" cy="8.5" r="1" />
    </template>
    <template v-else-if="name === 'more'">
      <circle cx="5" cy="12" r="1" fill="currentColor" stroke="none" />
      <circle cx="12" cy="12" r="1" fill="currentColor" stroke="none" />
      <circle cx="19" cy="12" r="1" fill="currentColor" stroke="none" />
    </template>
  </svg>
</template>
