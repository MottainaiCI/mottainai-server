import logo from "@/assets/images/logo.png"
import { route } from "preact-router"
import { useContext } from "preact/hooks"

import ThemeContext, { THEME_OPTIONS } from "@/contexts/theme"
import UserContext from "@/contexts/user"
import themes from "@/themes"
import UserService from "@/service/user"

import { SidebarItem, SidebarLink } from "./common"
import { SidebarPopoutSelector } from "./popout_selector"

const SignOut = () => {
  let { setUser } = useContext(UserContext)
  const signOut = () => {
    UserService.logout().then(() => {
      setUser(null)
      route("/")
    })
  }

  return (
    <SidebarItem
      icon="sign-out-alt"
      className="cursor-pointer"
      onClick={signOut}
    >
      Log out
    </SidebarItem>
  )
}

const Sidebar = () => {
  let { theme, setTheme } = useContext(ThemeContext)
  let { user } = useContext(UserContext)

  return (
    <div
      className={`flex-1 flex flex-col ${themes[theme].sidebar.bg} ${themes[theme].sidebar.bg}`}
    >
      <div className="flex flex-row justify-center items-center py-4">
        <img src={logo} className="w-10 mr-2" />
        <div className="text-2xl font-medium">MottainaiCI</div>
      </div>
      <div className="border h-px w-4/5 mx-auto mb-4" />
      <div className="flex-1 flex flex-col justify-between">
        <div className="flex flex-col">
          <SidebarLink href="/" icon="tachometer-alt" text="Dashboard" />
          {user && (
            <>
              <SidebarLink href="/tasks" icon="tasks" text="Tasks" />
              <SidebarLink href="/plans" icon="clock" text="Plans" />
              <SidebarLink
                href="/pipelines"
                icon="code-branch"
                text="Pipelines"
              />
              <SidebarLink href="/artefacts" icon="cloud" text="Artefacts" />
            </>
          )}
        </div>

        <div className="flex flex-col">
          {user ? (
            <SignOut />
          ) : (
            <SidebarLink href="/login" icon="user" text="Log In" />
          )}
          <SidebarPopoutSelector
            anchor="bottom"
            label="Theme"
            options={THEME_OPTIONS}
            onSelect={setTheme}
            selected={theme}
          />
        </div>
      </div>
    </div>
  )
}

export default Sidebar
