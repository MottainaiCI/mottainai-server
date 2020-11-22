import { createContext } from "preact"

const THEME_OPTIONS = [
  {
    label: "Mottainai Light",
    value: "mott-light",
  },
  {
    label: "Mottainai Dark",
    value: "mott-dark",
  },
]

const Theme = createContext({
  theme: THEME_OPTIONS[0].value,
  setTheme() {},
})

export { THEME_OPTIONS }
export default Theme
