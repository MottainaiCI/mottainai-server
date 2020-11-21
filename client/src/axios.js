import axios from "axios"

const instance = axios.create({
  baseURL: `${location.origin}/api/v1/client/`,
})

instance.interceptors.request.use((config) => {
  const csrf = localStorage.getItem("mottainai:csrf")
  if (csrf) {
    config.headers = {
      "X-CSRFToken": csrf,
    }
  }
  return config
}, Promise.reject)

instance.interceptors.response.use((response) => {
  localStorage.setItem("mottainai:csrf", response.headers["x-csrftoken"])
  return response
})

export default instance
