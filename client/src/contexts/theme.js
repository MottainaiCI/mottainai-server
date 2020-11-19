import { createContext } from 'preact'
const Theme = createContext({
  theme: 'light',
  setTheme(){}
})
export default Theme
