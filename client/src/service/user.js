import axios from "@/axios"

const UserService = {
  login(username, password) {
    return new Promise((res, rej) => {
      axios.post("/auth/login", { username, password }).then(({ data }) => {
        localStorage.setItem("mottainai:auth", 1)
        res(data)
      }, rej)
    })
  },
  logout() {
    return new Promise((res, rej) => {
      axios.post("/auth/logout").then(() => {
        localStorage.removeItem("mottainai:auth")
        res()
      }, rej)
    })
  },
  getUser() {
    return axios.get("/auth/user").then(({ data }) => data)
  },
  clearUser() {
    localStorage.removeItem("mottainai:auth")
  },
  isLoggedIn() {
    return !!localStorage.getItem("mottainai:auth")
  },
}

export default UserService
