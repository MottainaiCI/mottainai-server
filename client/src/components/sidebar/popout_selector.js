import { SidebarItem } from "./common"
import { FontAwesomeIcon } from "@aduh95/preact-fontawesome"
import { useContext } from "preact/hooks"
import ThemeContext from "@/contexts/theme"
import themes from "@/themes"

const SidebarPopoutSelector = ({
  anchor = "top",
  label,
  options,
  selected,
  onSelect = () => {},
}) => {
  let { theme } = useContext(ThemeContext)
  const optionEl = options.map(({ label, value }) => (
    <li>
      <a
        className={`py-2 px-4 block whitespace-no-wrap
          ${
            value === selected
              ? themes[theme].dropdown.selected
              : themes[theme].dropdown.notSelected
          }`}
        href="#"
        onClick={() => onSelect(value)}
      >
        {label}
      </a>
    </li>
  ))
  return (
    <SidebarItem
      icon="palette"
      className={`cursor-pointer group relative hover:${themes[theme].sidebar.activeBg}`}
    >
      <div className="flex justify-between items-center">
        <div className="flex-none">{label}</div>
        <div className="flex-none">
          <FontAwesomeIcon icon="caret-right" className="mr-3" />
        </div>
      </div>
      <ul
        className={`absolute ${anchor}-0 left-full w-max hidden group-hover:block`}
      >
        {optionEl}
      </ul>
    </SidebarItem>
  )
}

export { SidebarPopoutSelector }
