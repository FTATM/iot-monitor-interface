<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h2>System Dashboard</h2>
      <button @click="saveLayout" class="save-btn">Save Layout</button>
    </div>

    <GridLayout
      v-model:layout="layout"
      :col-num="12"
      :row-height="30"
      is-draggable
      is-resizable
    >
      <GridItem
        v-for="item in layout"
        :key="item.i"
        :x="item.x"
        :y="item.y"
        :w="item.w"
        :h="item.h"
        :i="item.i"
      >
         <div class="widget-container">
           <!-- Vue dynamically resolves the component using the map -->
           <component :is="widgetMap[item.widgetType]" />
         </div>
      </GridItem>
    </GridLayout>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { GridLayout, GridItem } from 'grid-layout-plus';

// 1. Import your widget components
import RevenueChartWidget from '../components/widgets/RevenueChartWidget.vue';

// 2. Create a map matching the string in your layout data to the actual component
const widgetMap = {
  'RevenueChartWidget': RevenueChartWidget,
  // You will add more here later, e.g.:
  // 'ActiveUsersWidget': ActiveUsersWidget,
};

// 3. Your layout data
const layout = ref([
  // This will now render the graph component
  { x: 0, y: 0, w: 6, h: 8, i: '1', widgetType: 'RevenueChartWidget' }, 
  
  // This will render blank for now because we haven't built ActiveUsersWidget yet
  { x: 6, y: 0, w: 6, h: 8, i: '2', widgetType: 'ActiveUsersWidget' } 
]);

const saveLayout = () => {
  console.log('Sending to Go:', JSON.stringify(layout.value));
};
</script>

<style scoped>
.dashboard-page {
  height: 100%;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.save-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}
/* The container wrapping the dynamic component */
.widget-container {
  width: 100%;
  height: 100%;
  background-color: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden; /* Prevents graphs from spilling out */
  cursor: grab;
}
.widget-container:active {
  cursor: grabbing;
}
</style>