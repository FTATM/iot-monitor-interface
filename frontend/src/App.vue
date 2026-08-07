<template>
  <!-- If the route asks to hide the layout (like the Login page) -->
  <template v-if="$route.meta.hideLayout">
    <router-view />
  </template>

  <!-- Default behavior: Use the MainLayout (like the Dashboard) -->
  <template v-else>
    <MainLayout>
      <router-view />
    </MainLayout>
  </template>
</template>

<script setup>
import MainLayout from './layouts/MainLayout.vue';
</script>

<style>
/* Let Tailwind handle box-sizing and resets globally */
/* html, body {
  margin: 0;
  padding: 0;
  width: 100vw;
  height: 100vh;
  overflow: hidden; 
  font-family: sans-serif;
} */

#app {
  width: 100%;
  height: 100%;
}
</style>