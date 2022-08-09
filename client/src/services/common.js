import axios from '.';

const api = () => axios();

const CommonService = {
    fetchTodos: () => api().get('/tasks/'),
    addTodo: (name) => api().post('/tasks/', { name, }),
    updateTask: (id, name, isDone) => api().put('/tasks/' + id, { name, isDone, }),
    deleteTask: (id) => api().delete('/tasks/' + id),
};

export default CommonService;
