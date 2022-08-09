import { createApp } from 'vue';
import App from './App.vue';
import router from '@/router';
import 'bootstrap';
import 'bootstrap/scss/bootstrap.scss';
import axios from 'axios';
import Toast from 'vue-toastification';
import 'vue-toastification/dist/index.css';

const app = createApp(App);
axios.defaults.baseURL = process.env.VUE_APP_SERVER_URL;
app.use(router);
app.use(Toast, {
    toastClassName: 'toast-body',
    timeout: 4000,
})
app.mount('#app')
