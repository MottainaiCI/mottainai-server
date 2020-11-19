import logo from "@/assets/images/logo.png"
import { Link } from 'preact-router/match';
import { FontAwesomeIcon } from '@aduh95/preact-fontawesome'
import { useContext } from 'preact/hooks';

import ThemeContext from "@/contexts/theme"
import themes from '@/themes'


const SidebarItem = ({icon, children}) => {
  return (<div class="py-2 px-4">
    <span className="w-8 inline-block text-center">
      {icon && <FontAwesomeIcon icon={icon} />}
    </span>
    <span className="ml-2 text-lg">{children}</span>
  </div>)
}

const SidebarLink = ({icon, text, ...props}) => {
  let {theme} = useContext(ThemeContext)
  return (<Link class="py-2 px-4" activeClassName={themes[theme].sidebar.activeBg} {...props}>
    <span className="w-8 inline-block text-center">
      <FontAwesomeIcon icon={icon} />
    </span>
    <span className="ml-2 text-lg">
      {text}
    </span>
  </Link>)
}

const Sidebar = () => {
  let { theme, setTheme } = useContext(ThemeContext)

  function toggleTheme() {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }

  return (<div className={`flex-1 flex flex-col ${themes[theme].sidebar.bg} ${themes[theme].sidebar.bg}`}>
    <div className="flex flex-row justify-center items-center py-4">
      <img src={logo} className="w-10 mr-2" />
      <div className="text-2xl font-medium">MottainaiCI</div>
    </div>
    <div className="border h-px w-4/5 mx-auto mb-4" />
    <div className="flex-1 flex flex-col justify-between">
      <div className="flex flex-col">
        <SidebarLink href="/" icon="tachometer-alt" text="Dashboard" />
        <SidebarLink href="/tasks" icon="tasks" text="Tasks" />
        <SidebarLink href="/plans" icon="clock" text="Plans" />
        <SidebarLink href="/pipelines" icon="code-branch" text="Pipelines" />
        <SidebarLink href="/artefacts" icon="cloud" text="Artefacts" />
      </div>

      <div className="flex flex-col">
        <SidebarItem icon="palette">
          <button className="focus:outline-none" onClick={toggleTheme}>Toggle Theme</button>
        </SidebarItem>
      </div>
    </div>
  </div>)
}

export default Sidebar
