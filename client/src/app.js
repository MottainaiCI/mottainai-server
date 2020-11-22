import { Router, route, getCurrentUrl } from "preact-router"
import { useLocalStorage } from "@rehooks/local-storage"
import { useState } from "preact/hooks"

import Dashboard from "@/components/pages/dashboard"
import Plans from "@/components/pages/plans"
import Pipelines from "@/components/pages/pipelines"
import Tasks from "@/components/pages/tasks"
import Artefacts from "@/components/pages/artefacts"
import Login from "@/components/pages/login"
import Sidebar from "@/components/sidebar"
import Spinner from "@/components/spinner"

import ThemeContext, { THEME_OPTIONS } from "@/contexts/theme"
import UserContext from "@/contexts/user"
import themes from "@/themes"
import axios from "@/axios"
import UserService from "./service/user"

const AUTHED = ["/plans", "/pipelines", "/tasks", "/artefacts"]
const UNAUTHED = ["/login"]

const App = () => {
  const [theme, setTheme] = useLocalStorage(
    "mottainai-theme",
    THEME_OPTIONS[0].value
  )
  const themeValue = { theme, setTheme }
  const [user, setUser] = useState(null)
  const [loadingUser, setLoadingUser] = useState(true)
  const userVal = { user, setUser }

  axios.interceptors.response.use(
    (res) => res,
    (err) => {
      if (err.response.status == 403) {
        UserService.clearUser()
        setUser(null)
        handleRoute(getCurrentUrl())
      }
      throw err
    }
  )

  const handleRoute = (url) => {
    if (loadingUser) {
      return
    }
    if (
      (AUTHED.some((val) => url.startsWith(val)) && !user) ||
      (UNAUTHED.some((val) => url.startsWith(val)) && user)
    ) {
      route("/")
    }
  }

  function clearLoadingUser() {
    setTimeout(() => {
      setLoadingUser(false)
      handleRoute(getCurrentUrl())
    }, 500)
  }

  useState(() => {
    if (UserService.isLoggedIn()) {
      UserService.getUser().then(setUser).finally(clearLoadingUser)
    } else {
      clearLoadingUser()
    }
  })

  if (loadingUser) {
    return (
      <div
        className={`flex justify-center items-center min-h-screen w-full ${themes[theme].bg}`}
      >
        <Spinner />
      </div>
    )
  }

  return (
    <UserContext.Provider value={userVal}>
      <ThemeContext.Provider value={themeValue}>
        <div
          className={`flex h-screen ${themes[theme].bg} ${themes[theme].textColor}`}
        >
          <div className="w-60 flex-none flex flex-col">
            <Sidebar />
          </div>
          <div className="px-8 py-10 flex-1 overflow-auto">
            <Router onChange={(e) => handleRoute(e.url)}>
              <Dashboard path="/" />
              <Plans path="/plans" />
              <Pipelines path="/pipelines" />
              <Tasks path="/tasks" />
              <Artefacts path="/artefacts" />
              <Login path="/login" />
            </Router>
          </div>
        </div>
      </ThemeContext.Provider>
    </UserContext.Provider>
  )
}

export default App
