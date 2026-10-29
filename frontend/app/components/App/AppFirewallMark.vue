<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 42 })
</script>

<template>
  <svg
    class="firewall-mark"
    :width="size"
    :height="size"
    viewBox="0 0 48 48"
    role="img"
    aria-label="Great Firewall"
  >
    <defs>
      <linearGradient id="wall-glow" x1="7" y1="38" x2="42" y2="12">
        <stop offset="0" stop-color="#c33c2e" />
        <stop offset="1" stop-color="#ffb23e" />
      </linearGradient>
      <radialGradient id="gate-glow">
        <stop offset="0" stop-color="#fff9df" />
        <stop offset="0.45" stop-color="#ffb23e" />
        <stop offset="1" stop-color="#e05a3f" />
      </radialGradient>
    </defs>

    <rect
      class="firewall-mark__badge"
      x="3"
      y="3"
      width="42"
      height="42"
      rx="13"
    />
    <circle class="firewall-mark__sun" cx="36" cy="12" r="4" />
    <path class="firewall-mark__mountains" d="m7 27 8-10 5 5 7-9 10 14Z" />
    <path
      class="firewall-mark__wall-shadow"
      d="M5.5 36.5 13 31l7 2 8-7 6 1 8.5-8"
    />
    <path class="firewall-mark__wall" d="M5.5 33.5 13 28l7 2 8-7 6 1 8.5-8" />
    <path class="firewall-mark__tower" d="M15 20h14l-2.5 4H28v14H16V24h1.5Z" />
    <path class="firewall-mark__roof" d="m13 20 9-6 9 6Z" />
    <path
      class="firewall-mark__brick"
      d="M18 25h3v2h-3zm5 0h3v2h-3zM18 29h8v2h-8Z"
    />
    <path class="firewall-mark__gate" d="M20 38v-4a2 2 0 0 1 4 0v4Z" />
    <path class="firewall-mark__route" d="M7 36c8-1 11 5 19 2 6-2 8-8 15-9" />
    <circle class="firewall-mark__packet" cx="7" cy="36" r="1.7" />
    <circle class="firewall-mark__packet" cx="27" cy="37.5" r="1.7" />
    <circle class="firewall-mark__packet" cx="41" cy="29" r="1.7" />
  </svg>
</template>

<style scoped>
.firewall-mark {
  display: block;
  overflow: visible;
  filter: drop-shadow(3px 4px 0 rgba(var(--gfw-shadow), 0.12));
}
.firewall-mark__badge {
  fill: #2b211f;
}
.firewall-mark__sun {
  fill: #f7b13b;
  opacity: 0.92;
}
.firewall-mark__mountains {
  fill: #4a3834;
}
.firewall-mark__wall-shadow {
  fill: none;
  stroke: #170f0d;
  stroke-linecap: square;
  stroke-linejoin: bevel;
  stroke-width: 7;
}
.firewall-mark__wall {
  fill: none;
  stroke: url(#wall-glow);
  stroke-linecap: square;
  stroke-linejoin: bevel;
  stroke-width: 4;
}
.firewall-mark__tower {
  fill: #c33c2e;
}
.firewall-mark__roof {
  fill: #f7b13b;
}
.firewall-mark__brick {
  fill: #ffb38d;
  opacity: 0.86;
}
.firewall-mark__gate {
  fill: url(#gate-glow);
}
.firewall-mark__route {
  fill: none;
  stroke: #fff7e8;
  stroke-dasharray: 2.3 2.3;
  stroke-linecap: round;
  stroke-width: 1.2;
}
.firewall-mark__packet {
  fill: #fff7e8;
  stroke: #f7b13b;
  stroke-width: 0.8;
}
</style>
