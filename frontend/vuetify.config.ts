import { defineVuetifyConfiguration } from 'vuetify-nuxt-module/custom-configuration'
import { md3 } from 'vuetify/blueprints'

export default defineVuetifyConfiguration({
  blueprint: md3,
  icons: false,
  defaults: {
    VCard: { flat: true, border: true, rounded: 'lg' },
    VBtn: { rounded: 'lg' },
    VChip: { rounded: 'lg' },
    VDataTable: { fixedHeader: true, hover: true },
    VTextField: { color: 'primary', variant: 'outlined' },
    VSelect: { color: 'primary', variant: 'outlined' },
    VTextarea: { color: 'primary', variant: 'outlined' },
    VSwitch: { color: 'primary' },
  },
  display: { mobileBreakpoint: 'sm' },
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          background: '#F7F1E7',
          surface: '#FFF9F0',
          primary: '#C33C2E',
          secondary: '#2B211F',
          success: '#3F8C68',
          warning: '#D98E22',
          error: '#A92127',
          info: '#577A8A',
          'on-background': '#2B211F',
          'on-surface': '#2B211F',
          'surface-variant': '#F1E4D6',
        },
      },
    },
  },
})
