<template>
  <div class="chart-widget">
    <h3 class="title">Revenue (Last 7 Days)</h3>
    
    <!-- A simple mock bar chart using CSS Flexbox -->
    <div class="mock-chart">
      <div class="bar" style="height: 40%;"></div>
      <div class="bar" style="height: 70%;"></div>
      <div class="bar" style="height: 50%;"></div>
      <div class="bar" style="height: 90%;"></div>
      <div class="bar" style="height: 30%;"></div>
    </div>
  </div>
</template>

<script setup>
// Later, you will fetch real data from your Go API here
</script>

<style scoped>
.chart-widget {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
  background: white;
}

.title {
  margin: 0 0 16px 0;
  font-size: 1rem;
  color: #334155;
}

.mock-chart {
  flex: 1;
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  gap: 8px;
  border-bottom: 2px solid #e2e8f0;
  border-left: 2px solid #e2e8f0;
  padding: 8px;
}

.bar {
  width: 15%;
  background-color: #3b82f6;
  border-radius: 4px 4px 0 0;
  transition: height 0.3s;
}
</style>