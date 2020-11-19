const colors = require('tailwindcss/colors')

module.exports = {
  purge: [
    './src/**/*.js',
    './src/**/*.jsx',
  ],
  theme: {
    colors: {
      primary: 'var(--color-primary)',
      secondary: 'var(--color-secondary)',
      sidebar: 'var(--color-sidebar)',
      "sidebar-active": 'var(--color-sidebar-active)',
      accent: 'var(--color-accent)',
      white: colors.white,
    },
    fontFamily: {
      sans: ['Lato', 'sans-serif'],
    },
    extend: {}
  }
}
