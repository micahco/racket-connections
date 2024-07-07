/** @type {import('tailwindcss').Config} */

const colors = require('tailwindcss/colors')

module.exports = {
  content: ["./ui/html/**/*.html"],
  theme: {
    // https://communications.oregonstate.edu/brand-guide/visual-identity/colors
    colors: {
      black: colors.black,
      white: colors.white,
      stone: colors.stone,
      blue: colors.blue,
      red: colors.red,
      green: colors.green,
      yellow: colors.yellow,
      "beaver-orange": "#D73F09",
      "electric-beav": "#F7A162",
    },
  },
  plugins: [],
}

