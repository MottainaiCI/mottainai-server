import { useForm } from "react-hook-form"
import { useContext, useState } from "preact/hooks"
import { route } from "preact-router"

import UserContext from "@/contexts/user"
import ThemeContext from "@/contexts/theme"
import UserService from "@/service/user"
import themes from "@/themes"

const Login = () => {
  let [error, setError] = useState(null)
  let { setUser } = useContext(UserContext)
  let { theme } = useContext(ThemeContext)

  const { register, handleSubmit } = useForm()

  const onSubmit = ({ username, password }) => {
    if (username && password) {
      UserService.login(username, password).then(
        (data) => {
          setUser(data)
          route("/")
        },
        (err) => {
          setError(err.response.data.error)
        }
      )
    }
  }

  return (
    <div
      className={`rounded shadow rounded mx-auto w-min p-8 ${themes[theme].cardBg}  ${themes[theme].cardBorder}`}
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="mb-4">
          <label htmlFor="username" className="block">
            Username
          </label>
          <input
            name="username"
            className="text-cultured-black rounded border focus:outline-none focus:border-green-mottainai px-2 py-1"
            ref={register}
          />
        </div>

        <div className="mb-4">
          <label htmlFor="password" className="block">
            Password
          </label>
          <input
            type="password"
            name="password"
            autoComplete
            className="text-cultured-black rounded border focus:outline-none focus:border-green-mottainai px-2 py-1"
            ref={register}
          />
        </div>
        <button
          type="submit"
          className="focus:outline-none bg-green-mottainai text-white w-full rounded p-1"
        >
          Log In
        </button>
        {error && <div className="text-center mt-4 text-red-500">{error}</div>}
      </form>
    </div>
  )
}

export default Login
