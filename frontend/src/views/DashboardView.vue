<template>
  <div class="dashboard-wrapper">

    <button v-if="isEditMode && !isSidebarOpen" class="tool-toggle-btn" @click="isSidebarOpen = true">
      ✏️ Widgets
    </button>

    <aside class="tool-sidebar" :class="{
      'is-open': isEditMode && isSidebarOpen,
      'is-hidden-by-drag': isDraggingTool
    }">
      <div class="sidebar-header">
        <div>
          <h3 class="sidebar-title">Available Widgets</h3>
          <p class="sidebar-subtitle">Drag to canvas</p>
        </div>
        <button class="close-btn" @click="isSidebarOpen = false">✖</button>
      </div>

      <div class="tool-list">
        <div v-for="tool in availableTools" :key="tool.type" class="draggable-tool" draggable="true"
          @dragstart="onDragStart($event, tool.type)" @dragend="onDragEnd">
          <span class="tool-icon">{{ tool.icon }}</span>
          <span class="tool-name">{{ tool.name }}</span>
        </div>
      </div>
    </aside>

    <div class="dashboard-page">
      <div class="page-header">
        <div class="canvas-selector">
          <label for="canvas-select">Active Canvas:</label>
          <select id="canvas-select" v-model="activeCanvasId" @change="loadCurrentCanvas()"
            class="custom-select" :disabled="isEditMode">
            <!-- Iterate over the Map using [key, value] syntax -->
            <option v-for="[id, canvas] in allUserCanvasesMap" :key="id" :value="id">
              {{ canvas.name }}
            </option>
          </select>

          <span v-if="isEditMode" class="badge edit-badge">Editing Mode</span>
          <span v-else class="badge view-badge">View Only</span>
        </div>

        <div class="action-buttons">
          <template v-if="!isEditMode">
            <button @click="enterEditMode" class="btn btn-primary">
              ✏️ Edit Dashboard
            </button>
          </template>

          <template v-else>
            <button @click="cancelEdit" class="btn btn-secondary">
              Cancel
            </button>
            <button @click="saveCurrentCanvas" class="btn btn-success">
              💾 Save Changes
            </button>
          </template>
        </div>
      </div>

      <div class="grid-drop-zone" @dragover.prevent="onDragOver" @drop="onDrop">
        <GridLayout v-model:layout="activeLayout" :col-num="12" :row-height="30" :is-draggable="isEditMode"
          :is-resizable="isEditMode">
          <GridItem v-for="item in activeLayout" :key="item.i" :x="item.x" :y="item.y" :w="item.w" :h="item.h"
            :i="item.i">
            <div :class="[
              'widget-container',
              { 'is-editable': isEditMode },
              { 'is-ghost': item.i === 'drop-placeholder' }
            ]" @contextmenu.prevent="onRightClick($event, item.i)">
              <component :is="widgetMap[item.widgetType]" />
            </div>
          </GridItem>
        </GridLayout>
      </div>
    </div>
  </div>

  <!-- --- NEW: CONTEXT MENU --- -->
  <div v-if="contextMenu.show" class="context-menu-overlay" @click="closeContextMenu"
    @contextmenu.prevent="closeContextMenu"></div>

  <div v-if="contextMenu.show" class="context-menu" :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }">
    <ul class="menu-list">
      <li class="menu-item">⚙️ Configure Widget (Coming Soon)</li>
      <li class="menu-item danger" @click="promptDelete">🗑️ Delete Widget</li>
    </ul>
  </div>

  <!-- --- NEW: DELETE CONFIRMATION MODAL --- -->
  <div v-if="showDeleteConfirm" class="modal-overlay">
    <div class="modal-content">
      <h3 class="modal-title">Delete Widget</h3>
      <p class="modal-text">Are you sure you want to remove this widget from the canvas? You can drag it back later if
        needed.</p>
      <div class="modal-actions">
        <button @click="showDeleteConfirm = false" class="btn btn-secondary">No, Cancel</button>
        <button @click="confirmDelete" class="btn btn-danger">Yes, Delete</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { GridLayout, GridItem } from 'grid-layout-plus';
import { useFetch } from '../composables/useFetch'
import { useMutation } from '../composables/useMutation'

import BarChart from '../components/widgets/BarChart.vue';
import BulletChart from '../components/widgets/BulletChart.vue';
import GaugeChart from '../components/widgets/GaugeChart.vue';

const widgetMap = {
  'BarChart': BarChart,
  'BulletChart': BulletChart,
  'GaugeChart': GaugeChart,
};

// --- DRAG AND DROP AUTO-HIDE LOGIC ---
const availableTools = ref([
  { type: 'BarChart', name: 'Bar Chart', icon: '📊' },
  { type: 'BulletChart', name: 'Bullet Chart', icon: '🎯' },
  { type: 'GaugeChart', name: 'Gauge', icon: '⏱️' }
]);

const baseUrl = import.meta.env.VITE_API_BASE_URL

// --- NEW STATE VARIABLES ---
const isEditMode = ref(false);
const isSidebarOpen = ref(false);
const isDraggingTool = ref(false);
const draggedWidgetType = ref(null);

// --- NEW CONTEXT MENU & MODAL STATE ---
const contextMenu = ref({ show: false, x: 0, y: 0, widgetId: null });
const showDeleteConfirm = ref(false);
const widgetToDelete = ref(null);

const allUserCanvasesMap = ref(new Map());

const activeCanvasId = ref(null);
const activeLayout = ref([]);
const widgetTypeMasterMap = new Map();

//! test
const testCanvas = 1;
const testUser = 1;

const loadUserCanvas = async (userId) => {
  const {
    data: userAllCanvasFetchData,
    isLoading: isUserAllCanvasFetchLoading,
    error: userAllCanvasFetchError,
    execute: userAllCanvasFetchApi
  } = useFetch();
  
  await userAllCanvasFetchApi(`${baseUrl}/canvas/getallbyuserid/${userId}`);

  if (!userAllCanvasFetchError.value && userAllCanvasFetchData.value) {
    // 1. Create a temporary standard JS Map
    const newMap = new Map();

    // 2. Loop through the API response and populate the Map
    userAllCanvasFetchData.value.forEach(canvas => {
      
      const formattedLayout = (canvas.widgets || []).map((widget, index) => {
        const typeInfo = widgetTypeMasterMap.get(widget.widgetTypeId);
        return {
          x: widget.layoutData.x,
          y: widget.layoutData.y,
          w: widget.layoutData.w,
          h: widget.layoutData.h,
          i: widget.widgetId ? widget.widgetId.toString() : index.toString(),
          widgetType: typeInfo ? typeInfo.widgetTypeName : ''
        };
      });

      // 3. Set the Map entry using the canvasId as the strict Key
      const stringId = canvas.canvasId.toString();
      newMap.set(stringId, {
        id: stringId,
        name: canvas.canvasName,
        layout: formattedLayout
      });
    });

    // 4. Assign the built map to our reactive Vue variable
    allUserCanvasesMap.value = newMap;

    // 5. Automatically load the first canvas to display on the screen
    if (newMap.size > 0) {
      // .keys().next().value is how you grab the first key from a Map
      activeCanvasId.value = newMap.keys().next().value; 
      loadCurrentCanvas(); 
    }
    
  } else if (userAllCanvasFetchError.value) {
    console.error("Failed to load user canvases:", userAllCanvasFetchError.value);
  }
};

const loadCurrentCanvas = () => {
  const selectedCanvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  
  if (selectedCanvas) {
    activeLayout.value = JSON.parse(JSON.stringify(selectedCanvas.layout));
  }
};

const setupData = async () => {
  const {
    data: widgetTypeMasterData,
    isLoading: isWidgetTypeMasterLoading,
    error: widgetTypeMasterError,
    execute: widgetTypeMasterApi
  } = useFetch()

  await widgetTypeMasterApi(`${baseUrl}/widgettype/getall`);

  // FIX 1: We want to run this if there is NO error, and data exists!
  if (!widgetTypeMasterError.value && widgetTypeMasterData.value) {

    // FIX 2: We must use .value to access the actual array returned by the API
    for (let i of widgetTypeMasterData.value) {
      widgetTypeMasterMap.set(i.widgetTypeId, i)
    }

  } else if (widgetTypeMasterError.value) {
    console.error("Failed to load widget types:", widgetTypeMasterError.value)
  }
}

const enterEditMode = () => {
  isEditMode.value = true;
  isSidebarOpen.value = false; // Show the pencil button first
};

const cancelEdit = () => {
  loadCurrentCanvas();
  isEditMode.value = false;
  isSidebarOpen.value = false;
};

const saveCurrentCanvas = async () => {
  const canvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  
  if (canvas) {
    // Update the layout inside the Map
    canvas.layout = JSON.parse(JSON.stringify(activeLayout.value));
    
    // Explicitly set it back to trigger Vue's reactivity for Maps
    allUserCanvasesMap.value.set(activeCanvasId.value, canvas);
  }
  
  isEditMode.value = false;
  isSidebarOpen.value = false;
  console.log("Current Map Data:", allUserCanvasesMap.value);
};


const onDragStart = (event, widgetType) => {
  event.dataTransfer.setData('widgetType', widgetType);
  event.dataTransfer.effectAllowed = 'copy';

  draggedWidgetType.value = widgetType; // Save what we are dragging for the preview

  setTimeout(() => {
    isDraggingTool.value = true;
  }, 0);
};

const onDragOver = (event) => {
  // Only run if we are actively dragging a tool in edit mode
  if (!isEditMode.value || !isDraggingTool.value || !draggedWidgetType.value) return;

  const dropZone = event.currentTarget.getBoundingClientRect();
  const mouseX = event.clientX - dropZone.left;
  const mouseY = event.clientY - dropZone.top; // Removed scrollTop because the grid wrapper handles it now

  const colNum = 12;
  const rowHeight = 30;
  const margin = 10; // The default gap between widgets in grid-layout-plus
  const newWidgetWidth = 4;

  const colWidth = dropZone.width / colNum;

  // Add the margin to the rowHeight for perfect vertical snapping
  let gridX = Math.floor(mouseX / colWidth);
  let gridY = Math.floor(mouseY / (rowHeight + margin));

  if (gridX + newWidgetWidth > colNum) gridX = colNum - newWidgetWidth;
  if (gridX < 0) gridX = 0;
  if (gridY < 0) gridY = 0;

  // Check if our ghost placeholder already exists on the board
  const placeholderIndex = activeLayout.value.findIndex(item => item.i === 'drop-placeholder');

  if (placeholderIndex !== -1) {
    // CRITICAL PERFORMANCE FIX: 
    // Only update the reactive layout if the grid coordinates actually changed.
    // If we don't do this, Vue will try to re-render the grid 60 times a second!
    const ghost = activeLayout.value[placeholderIndex];
    if (ghost.x !== gridX || ghost.y !== gridY) {
      ghost.x = gridX;
      ghost.y = gridY;
    }
  } else {
    // Create the ghost placeholder the moment the mouse enters the grid
    activeLayout.value.push({
      x: gridX,
      y: gridY,
      w: newWidgetWidth,
      h: 8,
      i: 'drop-placeholder',
      widgetType: draggedWidgetType.value // Render the actual chart they picked up!
    });
  }
};

// Triggered the moment the user drops the item (or cancels the drag)
const onDragEnd = () => {
  isDraggingTool.value = false;
  draggedWidgetType.value = null;

  // Clean up the ghost if it still exists (meaning the user dragged it outside the drop zone)
  activeLayout.value = activeLayout.value.filter(item => item.i !== 'drop-placeholder');
};

const onDrop = (event) => {
  if (!isEditMode.value) return;

  // Find the ghost placeholder we were dragging around
  const placeholderIndex = activeLayout.value.findIndex(item => item.i === 'drop-placeholder');

  if (placeholderIndex !== -1) {
    // Solidify the widget by giving it a real, unique ID based on the timestamp
    activeLayout.value[placeholderIndex].i = Date.now().toString();
  }

  // Clean up
  draggedWidgetType.value = null;
  isDraggingTool.value = false;
};

const onRightClick = (event, widgetId) => {
  if (!isEditMode.value) return; // Only allow config in Edit Mode

  contextMenu.value = {
    show: true,
    x: event.clientX,
    y: event.clientY,
    widgetId: widgetId
  };
};

// 2. Close the menu
const closeContextMenu = () => {
  contextMenu.value.show = false;
};

// 3. Trigger the confirmation modal
const promptDelete = () => {
  widgetToDelete.value = contextMenu.value.widgetId;
  closeContextMenu();
  showDeleteConfirm.value = true;
};

// 4. Execute the deletion
const confirmDelete = () => {
  activeLayout.value = activeLayout.value.filter(item => item.i !== widgetToDelete.value);
  showDeleteConfirm.value = false;
  widgetToDelete.value = null;
};

onMounted(async () => {
  await setupData();
  await loadUserCanvas(testUser);
});
</script>

<style scoped>
/* App Layout */
.dashboard-wrapper {
  display: flex;
  height: 100vh;
  width: 100%;
  background-color: #f1f5f9;
  position: relative;
  overflow: hidden;
}

/* --- Floating Pencil Button --- */
.tool-toggle-btn {
  position: absolute;
  left: 20px;
  top: 100px;
  /* Adjust this to sit below your header */
  background-color: #3b82f6;
  color: white;
  border: none;
  padding: 12px 20px;
  border-radius: 30px;
  font-weight: bold;
  font-size: 1rem;
  cursor: pointer;
  z-index: 40;
  box-shadow: 0 4px 10px rgba(59, 130, 246, 0.3);
  transition: transform 0.2s, background-color 0.2s;
}

.tool-toggle-btn:hover {
  background-color: #2563eb;
  transform: scale(1.05);
}

/* --- Auto-Hiding Sidebar --- */
.tool-sidebar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 250px;
  background-color: #ffffff;
  border-right: 1px solid #e2e8f0;
  padding: 20px;
  display: flex;
  flex-direction: column;
  z-index: 50;
  /* box-shadow: 4px 0 15px rgba(0,0,0,0.1);  */

  /* Hidden by default, handles all transitions seamlessly */
  transform: translateX(-100%);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Opens the sidebar when the user clicks the pencil */
.tool-sidebar.is-open {
  transform: translateX(0);
}

/* Overrides the open state the moment they start dragging */
.tool-sidebar.is-hidden-by-drag {
  transform: translateX(-100%) !important;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 12px;
}

.sidebar-title {
  margin: 0;
  font-size: 1.1rem;
  color: #1e293b;
  font-weight: 700;
}

.sidebar-subtitle {
  margin: 4px 0 0 0;
  font-size: 0.85rem;
  color: #64748b;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  color: #94a3b8;
  cursor: pointer;
  padding: 0 4px;
}

.close-btn:hover {
  color: #ef4444;
}

.tool-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.draggable-tool {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #f8fafc;
  border: 1px dashed #cbd5e1;
  border-radius: 6px;
  cursor: grab;
  transition: all 0.2s;
}

.draggable-tool:hover {
  background-color: #f1f5f9;
  border-color: #3b82f6;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.draggable-tool:active {
  cursor: grabbing;
}

/* Main Dashboard Area */
.dashboard-page {
  flex: 1;
  width: 100%;
  padding: 24px;
  overflow-y: auto;
}

/* Header Styling */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  background-color: white;
  padding: 16px 24px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.canvas-selector {
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 500;
  color: #334155;
}

.custom-select {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 1rem;
  background-color: #f8fafc;
  cursor: pointer;
}

.custom-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Badges */
.badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.view-badge {
  background-color: #e2e8f0;
  color: #475569;
}

.edit-badge {
  background-color: #fef08a;
  color: #854d0e;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    opacity: 1;
  }

  50% {
    opacity: 0.7;
  }

  100% {
    opacity: 1;
  }
}

/* Buttons */
.action-buttons {
  display: flex;
  gap: 12px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn:hover {
  opacity: 0.9;
}

.btn-primary {
  background-color: #3b82f6;
  color: white;
}

.btn-secondary {
  background-color: #e2e8f0;
  color: #334155;
}

.btn-success {
  background-color: #10b981;
  color: white;
}

/* Grid Widget Container */
.widget-container {
  width: 100%;
  height: 100%;
  background-color: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  cursor: default;
}

.widget-container.is-editable {
  cursor: grab;
  border: 2px dashed #94a3b8;
}

.widget-container.is-editable:active {
  cursor: grabbing;
}

.widget-container.is-ghost {
  opacity: 0.6;
  border: 2px dashed #3b82f6;
  /* Blue dashed border for the preview */
  pointer-events: none;
  /* Prevents the ghost from interfering with mouse events */
  transform: scale(0.98);
  /* Slightly shrinks it to look like it's floating */
  transition: all 0.1s ease;
}

.grid-drop-zone {
  flex: 1;
  min-height: 500px;
  /* Ensures there is always a drop zone even on an empty canvas */
  width: 100%;
}

/* --- Context Menu Styles --- */
.context-menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99;
  /* Sits just below the menu to catch outside clicks */
}

.context-menu {
  position: fixed;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.15);
  width: 220px;
  z-index: 100;
  padding: 8px 0;
}

.menu-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.menu-item {
  padding: 10px 16px;
  font-size: 0.9rem;
  color: #334155;
  cursor: pointer;
  transition: background-color 0.2s;
}

.menu-item:hover {
  background-color: #f1f5f9;
}

.menu-item.danger {
  color: #ef4444;
}

.menu-item.danger:hover {
  background-color: #fef2f2;
}

/* --- Modal Styles --- */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(15, 23, 42, 0.5);
  /* Dark transparent background */
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 200;
}

.modal-content {
  background: white;
  padding: 24px;
  border-radius: 8px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-title {
  margin: 0 0 12px 0;
  color: #1e293b;
  font-size: 1.25rem;
}

.modal-text {
  margin: 0 0 24px 0;
  color: #64748b;
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-danger {
  background-color: #ef4444;
  color: white;
}

.btn-danger:hover {
  background-color: #dc2626;
}

:deep(.vgl-item__resizer) {
  /* Increase the size of the grab area (default is usually 20px) */
  width: 35px !important;
  height: 35px !important;

  /* Optional: Adjust position slightly if it overlaps your chart borders */
  right: 2px !important;
  bottom: 2px !important;

  /* Make sure the mouse clearly shows the resize arrow */
  cursor: se-resize !important;
  z-index: 10 !important;
}

/* Optional: Make it visually clearer where to grab when in Edit Mode */
.widget-container.is-editable:hover+ :deep(.vgl-item__resizer)::after {
  content: '';
  position: absolute;
  right: 4px;
  bottom: 4px;
  width: 12px;
  height: 12px;
  border-right: 3px solid #3b82f6;
  /* Dashboard blue */
  border-bottom: 3px solid #3b82f6;
  border-radius: 2px;
}
</style>