/** @type {import('tailwindcss').Config} */

const colors = require('tailwindcss/colors')

module.exports = {
  content: ["./ui/html/**/*.html"],
  theme: {
    // https://communications.oregonstate.edu/brand-guide/visual-identity/colors
    colors: {
      transparent: 'transparent',
      current: 'currentColor',
      black: colors.black,
      white: colors.white,
      stone: colors.stone,
      blue: colors.blue,
      purple: colors.purple,
      red: colors.red,
      green: colors.green,
      yellow: colors.yellow,
      orange: colors.orange,
      "beaver-orange": "#D73F09",
      "electric-beav": "#F7A162",
    },
    extend: {
      minWidth: {
        '112': '28rem',
      }
    },
  },
  plugins: [],
}

