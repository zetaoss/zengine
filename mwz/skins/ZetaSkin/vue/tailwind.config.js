/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    '../includes/views/**/*.php',
    './src/**/*.{vue,js,ts,jsx,tsx}',
    './common/**/*.{vue,js,ts,jsx,tsx}',
  ],
  darkMode: 'class',
  plugins: [
  ],
  theme: {
    extend: {
      colors: {
        'z-bg': 'var(--z-bg)',
        'z-border': 'var(--z-border)',
        'z-card': 'var(--z-card)',
        'z-head': 'var(--z-head)',
        'z-link': 'var(--z-link)',
        'z-text': 'var(--z-text)',
        'z-text2': 'var(--z-text2)',
      },
    },
    screens: {
      'md': '900px',
      'lg': '1200px',
    },
  },
}
