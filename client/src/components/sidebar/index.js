import logo from "@/assets/images/logo.png"
import { Link } from 'preact-router/match';
import { FontAwesomeIcon } from '@aduh95/preact-fontawesome'

const SidebarItem = ({icon, text, ...props}) => (
  <Link className="py-2 px-4" activeClassName="bg-sidebar-active" {...props}>
    <span className="w-8 inline-block text-center">
      <FontAwesomeIcon icon={icon} />
    </span>
    <span className="ml-2 text-lg">
      {text}
    </span>
  </Link>
)


const Sidebar = () => (
  <div className="w-60 bg-sidebar">
    <div className="flex flex-row justify-center items-center py-4">
      <img src={logo} className="w-10 mr-2" />
      <div className="text-2xl font-medium">MottainaiCI</div>
    </div>
    <div className="border-accent border h-px w-4/5 mx-auto mb-4" />
    <div className="flex flex-col">
      <SidebarItem href="/" icon="tachometer-alt" text="Dashboard" />
      <SidebarItem href="/tasks" icon="tasks" text="Tasks" />
      <SidebarItem href="/plans" icon="clock" text="Plans" />
      <SidebarItem href="/pipelines" icon="code-branch" text="Pipelines" />
      <SidebarItem href="/artefacts" icon="cloud" text="Artefacts" />
    </div>
  </div>
)

export default Sidebar
