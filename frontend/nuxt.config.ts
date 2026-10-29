import { aliases } from 'vuetify/iconsets/mdi'

const mdiIcons = Object.values(aliases).map((icon) =>
  String(icon).replace(/^mdi-/, 'mdi:'),
)

export default defineNuxtConfig({
  ssr: false,
  compatibilityDate: '2025-07-15',
  modules: [
    '@pinia/nuxt',
    '@vueuse/nuxt',
    'vuetify-nuxt-module',
    '@nuxt/icon',
    '@nuxt/eslint',
    '@nuxt/test-utils/module',
  ],
  css: ['~/assets/styles/index.css'],
  typescript: { strict: true, shim: false },
  experimental: { typedPages: true },
  vuetify: {
    moduleOptions: { prefixComposables: true },
    vuetifyOptions: './vuetify.config.ts',
  },
  icon: { clientBundle: { icons: mdiIcons, scan: true } },
  runtimeConfig: { public: { apiBaseUrl: '' } },
})
