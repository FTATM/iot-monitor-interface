<template>
  <div class="app-wrapper">
    <!-- Sticky Top Header -->
    <header class="top-header">
      <div class="header-left">
        <button class="menu-toggle" @click="toggleSidebar">☰</button>
        
        <!-- Header Link to Dashboard -->
        <router-link to="/" class="logo">My App</router-link>
      </div>
      <div class="user-profile">Admin</div>
    </header>

    <div class="main-body">
      <!-- Sidebar Navigation -->
      <aside class="sidebar" :class="{ 'is-hidden': !isSidebarOpen }">
        <nav>
          <!-- Sidebar Links (active-class automatically applies when on that route) -->
          <router-link to="/" class="nav-item" active-class="active">Dashboard</router-link>
          <router-link to="/about" class="nav-item" active-class="active">About</router-link>
        </nav>
      </aside>

      <!-- Dynamic Content Area -->
      <main class="content-area">
        <slot /> 
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';

// State to track if the sidebar is open or closed
// Let's set it to false by default so it starts hidden, or true if you prefer!
const isSidebarOpen = ref(false);

const toggleSidebar = () => {
  isSidebarOpen.value = !isSidebarOpen.value;
};
</script>

<style scoped>
/* App Wrapper */
.app-wrapper {
  display: flex;
  flex-direction: column;
  height: 100vh; 
  overflow: hidden;
}

/* Header Styles */
.top-header {
  height: 60px;
  background-color: #1e293b;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.menu-toggle {
  background: transparent;
  border: none;
  color: white;
  font-size: 24px;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.menu-toggle:hover {
  background-color: #334155;
}

.logo {
  font-weight: bold;
  font-size: 1.2rem;
  text-decoration: none; 
  color: white;
}

/* Main Body Container */
.main-body {
  display: flex;
  flex: 1; 
  overflow: hidden; 
  position: relative; /* CRITICAL: This allows the absolute sidebar to position itself relative to this container */
}

/* Sidebar overlaying the content */
.sidebar {
  position: absolute; /* CRITICAL: Takes it out of document flow so it doesn't push content */
  top: 0;
  bottom: 0; /* Stretches it to the bottom of the main-body */
  left: 0;
  z-index: 50; /* Ensures it sits on top of the dashboard content */
  
  width: 250px;
  background-color: #f8fafc;
  border-right: 1px solid #e2e8f0;
  padding: 20px 0;
  box-shadow: 4px 0 10px rgba(0, 0, 0, 0.1); /* Adds a nice shadow over the content */
  
  /* Smoothly translates its X position */
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  
  overflow-y: auto;
}

/* Hidden state moves the sidebar 100% of its width to the left (off-screen) */
.sidebar.is-hidden {
  transform: translateX(-100%);
}

/* Nav Items */
.nav-item {
  display: block;
  padding: 12px 20px;
  color: #334155;
  text-decoration: none;
  font-weight: 500;
}

.nav-item:hover, .nav-item.active {
  background-color: #e2e8f0;
  color: #0f172a;
}

/* Content Area */
.content-area {
  flex: 1; 
  background-color: #f1f5f9;
  padding: 24px;
  overflow-y: auto; 
  width: 100%; /* Ensures content stays full width behind the sidebar */
}
</style>