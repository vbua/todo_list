import { createRouter, createWebHistory } from 'vue-router';
import TodoList from '../pages/TodoList.vue'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'todolist',
            component: TodoList
        },
        {
            path: '/:pathMatch(.*)*',
            name: 'NotFound',
            component: () => import('../pages/NotFound.vue')
        }
    ],
});
export default router;