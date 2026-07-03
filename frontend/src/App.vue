<template>
  <MainLayout>
    <!-- Vue Router will inject DashboardView or AboutView right here -->
    <router-view />
  </MainLayout>
</template>

<script setup>
import MainLayout from './layouts/MainLayout.vue';
</script>

<style>
/* Keep your global styles (html, body, #app resetting) here as before */
html, body {
  margin: 0;
  padding: 0;
  width: 100vw;
  height: 100vh;
  box-sizing: border-box;
  overflow: hidden; 
  font-family: sans-serif;
}
#app {
  width: 100%;
  height: 100%;
}
*, *::before, *::after {
  box-sizing: inherit;
}
</style>