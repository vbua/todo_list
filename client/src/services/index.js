import axios from 'axios'
const api = (url) => {
    const config = {
        baseURL: url || process.env.VUE_APP_SERVER_URL,
    }

    return axios.create(config)
}

export default api