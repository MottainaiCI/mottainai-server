const colors = require("tailwindcss/colors")

module.exports = {
  purge: ["./src/**/*.js", "./src/**/*.jsx"],
  theme: {
    extend: {
      colors: {
        beige: {
          100: "#f6f5f2",
          300: "#d9dcbf",
          500: "#908666",
          600: "#67592c",
          700: "#5d4530",
          750: "#463322",
          751: "#402915",
        },
        cultured: {
          white: "#fbfaf9",
          black: "#404547",
        },
        green: {
          mottainai: "#73ba25",
        },
      },
      fontFamily: {
        sans: ["Lato", "sans-serif"],
      },
    },
  },
}
