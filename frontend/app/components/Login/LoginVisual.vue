<script setup lang="ts">
/* eslint-disable vue/html-self-closing */
import loginGreatWall from '~/assets/images/login-great-wall.jpg'

defineProps<{ statusLabel: string; statusTone: 'online' | 'unavailable' }>()
</script>

<template>
  <section class="login-visual" aria-hidden="true">
    <img
      :src="loginGreatWall"
      alt=""
      class="login-visual__image"
      decoding="async"
      fetchpriority="high"
      height="1254"
      width="1254"
    />
    <div class="login-visual__wash" />
    <div class="login-visual__status" :class="`is-${statusTone}`">
      <span class="login-visual__status-ping" />
      <span>{{ statusLabel }}</span>
    </div>
    <span class="login-visual__sticker">Packets need clearance</span>
    <div class="login-visual__caption">
      <span class="login-visual__eyebrow"
        >Ancient perimeter. Modern policy.</span
      >
      <strong>One wall.<br />Every trusted route.</strong>
    </div>
  </section>
</template>

<style scoped>
.login-visual {
  position: relative;
  min-height: 540px;
  overflow: hidden;
  background: rgb(var(--gfw-soot));
}
.login-visual__image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 48%;
  animation: visual-arrive 720ms cubic-bezier(0.2, 0.8, 0.2, 1) both;
}
.login-visual__wash {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(24, 15, 14, 0.04) 30%, rgba(24, 15, 14, 0.88)),
    linear-gradient(90deg, rgba(24, 15, 14, 0.36), transparent 68%);
}
.login-visual__status,
.login-visual__caption,
.login-visual__sticker {
  position: absolute;
  z-index: 1;
}
.login-visual__status {
  top: 22px;
  left: 22px;
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 9px 13px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  border-radius: 999px;
  background: rgba(255, 249, 240, 0.92);
  box-shadow: 0 5px 0 rgba(43, 33, 31, 0.25);
  color: rgb(var(--gfw-text));
  font-size: 0.72rem;
  font-weight: 850;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  backdrop-filter: blur(12px);
}
.login-visual__status-ping {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #c33c2e;
  box-shadow: 0 0 0 0 rgba(195, 60, 46, 0.55);
  animation: status-ping 2.4s ease-out infinite;
}
.login-visual__status.is-unavailable .login-visual__status-ping {
  background: rgb(var(--v-theme-error));
  animation: none;
}
.login-visual__sticker {
  top: 27%;
  right: -25px;
  padding: 10px 32px 10px 17px;
  transform: rotate(5deg);
  border: 2px solid rgba(43, 33, 31, 0.3);
  border-radius: 11px 0 0 11px;
  background: #f7b13b;
  box-shadow: 0 7px 0 rgba(43, 33, 31, 0.32);
  color: #2b211f;
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}
.login-visual__caption {
  right: clamp(22px, 4vw, 44px);
  bottom: clamp(24px, 4vw, 42px);
  left: clamp(22px, 4vw, 44px);
  display: grid;
  gap: 8px;
  color: white;
  text-shadow: 0 3px 24px rgba(0, 0, 0, 0.55);
}
.login-visual__caption strong {
  font-size: clamp(1.9rem, 3.4vw, 3.15rem);
  line-height: 0.98;
  letter-spacing: -0.045em;
}
.login-visual__eyebrow {
  font: 800 0.72rem var(--gfw-mono);
  letter-spacing: 0.14em;
  text-transform: uppercase;
}
@keyframes visual-arrive {
  from {
    opacity: 0;
    transform: scale(1.035);
  }
}
@keyframes status-ping {
  55%,
  100% {
    box-shadow: 0 0 0 9px rgba(195, 60, 46, 0);
  }
}
@media (max-width: 960px) {
  .login-visual {
    min-height: clamp(260px, 42vw, 370px);
  }
}
@media (max-width: 600px) {
  .login-visual {
    min-height: 230px;
  }
  .login-visual__status {
    top: 14px;
    left: 14px;
  }
  .login-visual__sticker {
    display: none;
  }
  .login-visual__caption {
    right: 20px;
    bottom: 20px;
    left: 20px;
  }
  .login-visual__caption strong {
    font-size: 1.7rem;
  }
}
@media (prefers-reduced-motion: reduce) {
  .login-visual__image,
  .login-visual__status-ping {
    animation: none;
  }
}
</style>
