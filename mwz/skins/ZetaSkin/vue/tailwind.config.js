/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    '../templates/*.mustache',
    './src/**/*.{vue,js,ts,jsx,tsx}',
    './common/**/*.{vue,js,ts,jsx,tsx}',
  ],
  darkMode: 'class',
  theme: {
    screens: {
      'md': '900px',
      'lg': '1200px',
    },
  },
}
