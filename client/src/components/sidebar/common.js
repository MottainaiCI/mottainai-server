import { FontAwesomeIcon } from "@aduh95/preact-fontawesome"
import { Link } from "preact-router/match"
import { useContext } from "preact/hooks"
import ThemeContext from "@/contexts/theme"
import themes from "@/themes"

const SidebarItem = ({ icon, children, className = "", ...props }) => {
  return (
    <div className={`py-2 pl-4 flex ${className}`} {...props}>
      <div className="flex-none w-8 inline-block text-center">
        {icon && <FontAwesomeIcon icon={icon} />}
      </div>
      <div className="flex-1 ml-2 text-lg">{children}</div>
    </div>
  )
}

const SidebarLink = ({ icon, text, ...props }) => {
  let { theme } = useContext(ThemeContext)
  return (
    <Link
      class="py-2 pl-4 flex"
      activeClassName={themes[theme].sidebar.activeBg}
      {...props}
    >
      <div className="flex-none w-8 inline-block text-center">
        <FontAwesomeIcon icon={icon} />
      </div>
      <div className="flex-1 ml-2 text-lg">{text}</div>
    </Link>
  )
}

export { SidebarItem, SidebarLink }
