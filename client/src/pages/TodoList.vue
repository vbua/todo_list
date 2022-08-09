<script setup>
import {onMounted, ref} from "vue";
import CommonService from '../services/common';
import { useToast } from 'vue-toastification';

const toast = useToast();

const todos = ref([]);

const fetchTodos = async () => {
  const response = await CommonService.fetchTodos();
  if (response.data.error) {
    return toast.error(response.data.message);
  }
  todos.value = response.data.message;
};

onMounted(() => {
  fetchTodos();
});

const addTodo = async (submitEvent) => {
  if (!submitEvent.target.elements.name.value.length) {
    return toast.error("Enter name of your task.");
  }
  await CommonService.addTodo(submitEvent.target.elements.name.value);
  submitEvent.target.elements.name.value = "";
  await fetchTodos();
};

const finishTask = async (id, name) => {
  await CommonService.updateTask(id, name, true);
  await fetchTodos();
}

const deleteTask = async (id) => {
  await CommonService.deleteTask(id);
  await fetchTodos();
}
</script>
<template>
  <div class="d-flex align-items-center justify-content-center">
    <div>
      <h1 class="text-center my-3 pb-3">To Do App</h1>
      <form @submit.prevent="addTodo" class="row row-cols-lg-auto g-3 justify-content-center align-items-center mb-4 pb-2">
        <div>
          <input type="text" class="form-control" name="name" placeholder="Enter a task here">
        </div>
        <div><button type="submit" class="btn btn-primary">Add</button></div>
        <div></div>
      </form>
      <table class="table align-middle">
        <thead>
          <tr>
            <th scope="col">#</th>
            <th scope="col">Name</th>
            <th scope="col">Status</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="todo in todos" :key="todo.id">
            <td>{{ todo.id }}</td>
            <td>{{ todo.name }}</td>
            <td>{{ todo.isDone ? "Done" : "In process" }}</td>
            <td>
              <button class="btn btn-success me-2" @click.prevent="finishTask(todo.id, todo.name)">Finish</button>
              <button class="btn btn-danger" @click.prevent="deleteTask(todo.id)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
<style lang="scss" scoped>

</style>