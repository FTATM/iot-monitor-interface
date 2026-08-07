<template>
  <NoAccess v-if="!hasPermission('Canvas Design', 'Display')" />
  <div v-else class="flex h-screen w-full bg-base-200 relative overflow-hidden">

    <button v-if="!isSidebarOpen"
      class="btn btn-primary shadow-lg absolute left-5 top-[100px] z-40 rounded-full transition-transform hover:scale-105"
      @click="isSidebarOpen = true">
      <Icon icon="lucide:layout-grid" class="w-5 h-5 mr-1" /> Widgets
    </button>

    <!-- Sidebar -->
    <aside
      class="absolute left-0 top-0 bottom-0 w-[250px] bg-base-100 border-r border-base-300 p-5 flex flex-col z-50 transition-transform duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
      :class="[isSidebarOpen ? 'translate-x-0' : '-translate-x-full', isDraggingTool ? '!-translate-x-full' : '']">
      <div class="flex justify-between items-start mb-5 border-b-2 border-base-200 pb-3">
        <div>
          <h3 class="m-0 text-[1.1rem] font-bold text-base-content">Available Widgets</h3>
          <p class="mt-1 mb-0 text-[0.85rem] text-base-content/60">Drag to canvas</p>
        </div>
        <button class="btn btn-ghost btn-sm btn-square text-base-content/50 hover:text-error"
          @click="isSidebarOpen = false">
          <Icon icon="lucide:x" class="w-5 h-5" />
        </button>
      </div>

      <div class="flex flex-col gap-3">
        <div v-for="tool in availableTools" :key="tool.type"
          class="flex items-center gap-3 p-3 bg-base-200/50 border border-dashed border-base-300 rounded-box cursor-grab transition-all duration-200 hover:bg-base-200 hover:border-primary hover:shadow-sm active:cursor-grabbing"
          draggable="true" @dragstart="onDragStart($event, tool.type)" @dragend="onDragEnd">
          <Icon :icon="tool.icon" class="w-6 h-6 text-primary shrink-0" />
          <span class="text-sm font-medium text-base-content">{{ tool.name }}</span>
        </div>
      </div>
    </aside>

    <div class="flex-1 w-full p-6 overflow-y-auto">

      <!-- Top Control Bar (Editor Mode) -->
      <div
        class="flex flex-wrap md:flex-nowrap justify-between items-center gap-4 mb-6 bg-base-100 p-4 rounded-box shadow-sm border border-base-200">
        <div class="flex items-center gap-3 font-medium text-base-content overflow-x-auto">
          <label for="canvas-select" class="whitespace-nowrap">Editing Canvas:</label>
          <select id="canvas-select" v-model="activeCanvasId" @change="loadCurrentCanvas()"
            class="select select-bordered select-sm min-w-[200px]">
            <option v-for="[id, canvas] in allUserCanvasesMap" :key="id" :value="id">
              {{ canvas.name }}
            </option>
          </select>
          <span
            class="badge badge-warning badge-sm font-bold uppercase tracking-wider animate-pulse py-3 px-3 whitespace-nowrap shrink-0">
            Editor Mode
          </span>
        </div>

        <div class="flex gap-3 shrink-0">
          <button @click="loadCurrentCanvas" class="btn btn-ghost btn-sm md:btn-md">Reset Layout</button>
          <button @click="saveCurrentCanvas" class="btn btn-success text-white btn-sm md:btn-md">
            <Icon icon="lucide:save" class="w-4 h-4 mr-2" /> Save Changes
          </button>
        </div>
      </div>

      <!-- Editable Grid Layout -->
      <div class="flex-1 min-h-[500px] w-full" @dragover.prevent="onDragOver" @drop="onDrop">
        <GridLayout v-model:layout="activeLayout" :col-num="12" :row-height="30" :is-draggable="true"
          :is-resizable="true">
          <GridItem v-for="item in activeLayout" :key="item.i" :x="item.x" :y="item.y" :w="item.w" :h="item.h"
            :i="item.i">
            <div :class="[
              'w-full h-full bg-base-100 border border-base-200 rounded-box shadow-sm overflow-hidden flex flex-col',
              '!cursor-grab border-2 border-dashed !border-indigo-500/100 active:!cursor-grabbing is-editable',
              { 'opacity-60 border-2 border-dashed !border-primary pointer-events-none scale-[0.98]': item.i === 'drop-placeholder' }
            ]" @contextmenu.prevent="onRightClick($event, item.i)">
              <component :is="widgetMap[item.widgetTypeName]" :widget-data="item" />
            </div>
          </GridItem>
        </GridLayout>
      </div>
    </div>
  </div>

  <!-- Context Menu -->
  <template v-if="hasPermission('Canvas Design', 'Display')">
    <div v-if="contextMenu.show" class="fixed inset-0 z-[99]" @click="closeContextMenu"
      @contextmenu.prevent="closeContextMenu"></div>
    <div v-if="contextMenu.show" class="fixed z-[100]"
      :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }">
      <ul class="menu bg-base-100 w-56 rounded-box shadow-xl border border-base-200">
        <li>
          <a class="hover:bg-base-200" @click="promptConfig">
            <Icon icon="lucide:settings" class="w-4 h-4 text-base-content/70" /> Configure Widget
          </a>
        </li>
        <li>
          <a class="text-error hover:bg-error/10 hover:text-error" @click="promptDelete">
            <Icon icon="lucide:trash-2" class="w-4 h-4" /> Delete Widget
          </a>
        </li>
      </ul>
    </div>

    <!-- Delete Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box modal-animate">
        <h3 class="font-bold text-lg text-base-content">Delete Widget</h3>
        <p class="py-4 text-base-content/70">Are you sure you want to remove this widget from the canvas?</p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost">No, Cancel</button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white">Yes, Delete</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
    </dialog>

    <!-- Config Modal -->
    <dialog ref="configModal" class="modal z-[200]">
      <div class="modal-box max-w-2xl min-h-[450px] flex flex-col p-0">
        <div class="px-6 pt-6 pb-2 relative">
          <button
            class="btn btn-sm btn-circle btn-ghost absolute right-4 top-4 text-base-content/50 hover:text-error hover:bg-error/10"
            @click="closeConfigModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
          <h3 class="font-bold text-lg text-base-content flex items-center gap-2 pr-8">
            <Icon icon="lucide:settings-2" class="w-5 h-5 text-primary" /> Widget Configuration
            <span class="badge badge-neutral ml-auto">{{ configForm.widgetTypeName }}</span>
          </h3>
          <div role="tablist" class="tabs tabs-bordered mt-4 w-full">
            <a role="tab" class="tab h-10" :class="{ 'tab-active': activeTab === 'general' }"
              @click="activeTab = 'general'">
              <span class="font-semibold">General Settings</span>
            </a>
            <a v-if="widgetConfigComponentMap[configForm.widgetTypeName]" role="tab" class="tab h-10"
              :class="{ 'tab-active': activeTab === 'chart' }" @click="activeTab = 'chart'">
              <span class="font-semibold">Chart Configuration</span>
            </a>
          </div>
        </div>

        <div class="p-6 flex-1 overflow-y-auto bg-base-100/50">
          <div v-show="activeTab === 'general'" class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="flex flex-col gap-3">
              <label class="form-control w-full">
                <div class="label pb-1"><span class="label-text font-bold">Widget Label</span></div>
                <input type="text" v-model="configForm.widgetLabel" class="input input-bordered input-sm w-full" />
              </label>
              <label class="form-control w-full">
                <div class="label pb-1"><span class="label-text font-bold">Data Source (Device ID)</span></div>
                <select v-model="configForm.deviceId" class="select select-bordered select-sm w-full">
                  <option disabled :value="0">Select a device...</option>
                  <option :value="1">Main Server</option>
                  <option :value="2">Database Cluster</option>
                </select>
              </label>
              <label class="form-control w-full mt-2 items-center cursor-pointer">
                <div class="label pb-1 w-full"><span class="label-text font-bold">Panel Background Color</span></div>
                <input type="color" v-model="configForm.widgetColor.bgHex"
                  class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
              </label>
            </div>
            <div class="flex flex-col gap-3">
              <div class="label pb-0"><span class="label-text font-bold">Grid Position & Size</span></div>
              <div class="grid grid-cols-4 gap-2">
                <label class="form-control"><span class="label-text-alt mb-1 text-center font-semibold">X
                    (Col)</span><input type="number" v-model="configForm.layoutData.x"
                    class="input input-bordered input-sm w-full text-center" /></label>
                <label class="form-control"><span class="label-text-alt mb-1 text-center font-semibold">Y
                    (Row)</span><input type="number" v-model="configForm.layoutData.y"
                    class="input input-bordered input-sm w-full text-center" /></label>
                <label class="form-control"><span
                    class="label-text-alt mb-1 text-center font-semibold">Width</span><input type="number"
                    v-model="configForm.layoutData.w"
                    class="input input-bordered input-sm w-full text-center" /></label>
                <label class="form-control"><span
                    class="label-text-alt mb-1 text-center font-semibold">Height</span><input type="number"
                    v-model="configForm.layoutData.h"
                    class="input input-bordered input-sm w-full text-center" /></label>
              </div>
            </div>
          </div>
          <div v-if="widgetConfigComponentMap[configForm.widgetTypeName]" v-show="activeTab === 'chart'">
            <component :is="widgetConfigComponentMap[configForm.widgetTypeName]"
              :modelValue="configForm.customChartData" :key="widgetToConfig"
              @update:modelValue="(val) => configForm.customChartData = val" />
          </div>
        </div>
        <div class="px-6 py-4 border-t border-base-200 bg-base-100 flex justify-end gap-3 rounded-b-box">
          <button type="button" @click="closeConfigModal" class="btn btn-ghost">Cancel</button>
          <button type="button" @click="saveConfig" class="btn btn-primary text-white px-8">Save Settings</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeConfigModal">close</button></form>
    </dialog>
  </template>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { GridLayout, GridItem } from 'grid-layout-plus';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';
import { toast } from 'vue3-toastify';
import { usePermissionStore } from '@/stores/usePermissionStore';
import { Icon } from '@iconify/vue';

import NoAccess from '@/components/NoAccess.vue';
import BarChart from '@/components/widgets/BarChart.vue';
import BarChartConfig from '@/components/widgets/BarChartConfig.vue';
import BulletChart from '@/components/widgets/BulletChart.vue';
import BulletChartConfig from '@/components/widgets/BulletChartConfig.vue';
import GaugeChart from '@/components/widgets/GaugeChart.vue';
import GaugeChartConfig from '@/components/widgets/GaugeChartConfig.vue';
import LineChart from '@/components/widgets/LineChart.vue';
import LineChartConfig from '@/components/widgets/LineChartConfig.vue';
import PieChart from '@/components/widgets/PieChart.vue'
import PieChartConfig from '@/components/widgets/PieChartConfig.vue'
import ScatterChart from '@/components/widgets/ScatterChart.vue'
import ScatterChartConfig from '@/components/widgets/ScatterChartConfig.vue'
import BarProcess from '@/components/widgets/BarProcess.vue'
import BarProcessConfig from '@/components/widgets/BarProcessConfig.vue'
import Status from '@/components/widgets/Status.vue'
import StatusConfig from '@/components/widgets/StatusConfig.vue'
import Table from '@/components/widgets/Table.vue'
import TableConfig from '@/components/widgets/TableConfig.vue'
import Alert from '@/components/widgets/Alert.vue'
import AlertConfig from '@/components/widgets/AlertConfig.vue'
import Text from '@/components/widgets/Text.vue'
import TextConfig from '@/components/widgets/TextConfig.vue'

const widgetMap = {
  'BarChart': BarChart,
  'BulletChart': BulletChart,
  'GaugeChart': GaugeChart,
  'LineChart': LineChart,
  'PieChart': PieChart,
  'ScatterChart': ScatterChart,
  'BarProcess': BarProcess,
  'Status': Status,
  'Table': Table,
  'Alert': Alert,
  'Text': Text,
};

const widgetConfigComponentMap = {
  'BarChart': BarChartConfig,
  'GaugeChart': GaugeChartConfig,
  'BulletChart': BulletChartConfig,
  'LineChart': LineChartConfig,
  'PieChart': PieChartConfig,
  'ScatterChart': ScatterChartConfig,
  'BarProcess': BarProcessConfig,
  'Status': StatusConfig,
  'Table': TableConfig,
  'Alert': AlertConfig,
  'Text': TextConfig,
};

// --- STORE ---
const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const availableTools = ref([
  { type: 'BarChart', name: 'Bar Chart', icon: 'lucide:bar-chart-3' },
  { type: 'BulletChart', name: 'Bullet Chart', icon: 'lucide:target' },
  { type: 'GaugeChart', name: 'GaugeChart', icon: 'lucide:gauge' },
  { type: 'LineChart', name: 'Line Chart', icon: 'lucide:line-chart' },
  { type: 'PieChart', name: 'Pie Chart', icon: 'lucide:pie-chart' },
  { type: 'ScatterChart', name: 'Scatter Chart', icon: 'lucide:scatter-chart' },
  { type: 'BarProcess', name: 'Bar Process', icon: 'lucide:bar-chart-horizontal' },
  { type: 'Status', name: 'Status', icon: 'lucide:activity' },
  { type: 'Table', name: 'Table', icon: 'lucide:table' },
  { type: 'Alert', name: 'Alert', icon: 'lucide:bell' },
  { type: 'Text', name: 'Text', icon: 'lucide:type' },
]);

const activeTab = ref('general');
const isSidebarOpen = ref(false);
const isDraggingTool = ref(false);
const draggedWidgetTypeName = ref(null);

const contextMenu = ref({ show: false, x: 0, y: 0, widgetId: null });
const deleteModal = ref(null);
const configModal = ref(null);
const configForm = ref({ widgetLabel: '', deviceId: 0, widgetColor: { bgHex: '', textHex: '', chartHex: '' }, layoutData: { x: 0, y: 0, w: 4, h: 8 }, customChartData: {} });
const widgetToDelete = ref(null);
const widgetToConfig = ref(null);

const allUserCanvasesMap = ref(new Map());
const activeCanvasId = ref(null);
const activeLayout = ref([]);
const widgetTypeMasterMap = new Map();

const { data: widgetTypeMasterData, error: widgetTypeMasterError, execute: widgetTypeMasterApi } = useFetch();
const { data: userAllCanvasFetchData, error: userAllCanvasFetchError, execute: userAllCanvasFetchApi } = useFetch();
const { res: upsertRes, error: upsertError, execute: upsertWidgetApi } = useMutation();

const loadUserCanvas = async () => {
  await userAllCanvasFetchApi('/canvas/getalldetailbyuser');
  if (!userAllCanvasFetchError.value && userAllCanvasFetchData.value) {
    const newMap = new Map();
    userAllCanvasFetchData.value.data.forEach(canvas => {
      const formattedLayout = (canvas.widgets || []).map((widget, index) => {
        const typeInfo = widgetTypeMasterMap.get(widget.widgetTypeId);
        return {
          x: widget.layoutData.x, y: widget.layoutData.y, w: widget.layoutData.w, h: widget.layoutData.h,
          i: widget.widgetId ? widget.widgetId.toString() : index.toString(),
          widgetTypeName: typeInfo ? typeInfo.widgetTypeName : '',
          widgetId: widget.widgetId, widgetTypeId: widget.widgetTypeId, deviceId: widget.deviceId || 0,
          widgetLabel: widget.widgetLabel || '', widgetColor: widget.widgetColor || { bgHex: '', textHex: '', chartHex: '' },
          customChartData: widget.customChartData || {}
        };
      });
      const stringId = canvas.canvasId.toString();
      newMap.set(stringId, { id: stringId, name: canvas.canvasName, layout: formattedLayout });
    });
    allUserCanvasesMap.value = newMap;
    if (newMap.size > 0) {
      activeCanvasId.value = newMap.keys().next().value;
      loadCurrentCanvas();
    }
  } else if (userAllCanvasFetchError.value) {
    toast.error("Failed to load user canvas");
  }
};

const loadCurrentCanvas = () => {
  const selectedCanvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  if (selectedCanvas) {
    activeLayout.value = JSON.parse(JSON.stringify(selectedCanvas.layout));
  }
};

const setupData = async () => {
  await widgetTypeMasterApi('/widgettype/getall');
  if (!widgetTypeMasterError.value && widgetTypeMasterData.value) {
    for (let i of widgetTypeMasterData.value.data) widgetTypeMasterMap.set(i.widgetTypeId, i);
  } else if (widgetTypeMasterError.value) {
    toast.error("Failed to load widget types");
  }
};

const saveCurrentCanvas = async () => {
  const widgetsList = activeLayout.value.map(item => ({
    widgetId: item.widgetId || 0, widgetTypeId: item.widgetTypeId || 0, deviceId: item.deviceId || 0,
    widgetLabel: item.widgetLabel || 'New Widget', layoutData: { x: item.x, y: item.y, w: item.w, h: item.h },
    widgetColor: item.widgetColor, customChartData: item.customChartData
  }));
  const payload = { canvasId: Number(activeCanvasId.value), UpsertWidgets: widgetsList };
  await upsertWidgetApi('/widget/upsert', payload, 'POST');

  if (!upsertError.value && upsertRes.value?.ok) {
    const canvas = allUserCanvasesMap.value.get(activeCanvasId.value);
    if (canvas) {
      canvas.layout = JSON.parse(JSON.stringify(activeLayout.value));
      allUserCanvasesMap.value.set(activeCanvasId.value, canvas);
    }
    toast.success("Canvas saved successfully!");
  } else {
    toast.error("Failed to save canvas widgets");
  }
};

const onDragStart = (event, widgetTypeName) => {
  event.dataTransfer.setData('widgetTypeName', widgetTypeName);
  event.dataTransfer.effectAllowed = 'copy';
  draggedWidgetTypeName.value = widgetTypeName;
  setTimeout(() => isDraggingTool.value = true, 0);
};

const onDragOver = (event) => {
  if (!isDraggingTool.value || !draggedWidgetTypeName.value) return;
  const dropZone = event.currentTarget.getBoundingClientRect();
  const mouseX = event.clientX - dropZone.left;
  const mouseY = event.clientY - dropZone.top;
  const colNum = 12, rowHeight = 30, margin = 10, newWidgetWidth = 4;
  const colWidth = dropZone.width / colNum;
  let gridX = Math.floor(mouseX / colWidth);
  let gridY = Math.floor(mouseY / (rowHeight + margin));
  if (gridX + newWidgetWidth > colNum) gridX = colNum - newWidgetWidth;
  if (gridX < 0) gridX = 0;
  if (gridY < 0) gridY = 0;

  const placeholderIndex = activeLayout.value.findIndex(item => item.i === 'drop-placeholder');
  if (placeholderIndex !== -1) {
    const ghost = activeLayout.value[placeholderIndex];
    if (ghost.x !== gridX || ghost.y !== gridY) { ghost.x = gridX; ghost.y = gridY; }
  } else {
    activeLayout.value.push({ x: gridX, y: gridY, w: newWidgetWidth, h: 8, i: 'drop-placeholder', widgetTypeName: draggedWidgetTypeName.value });
  }
};

const onDragEnd = () => {
  isDraggingTool.value = false;
  draggedWidgetTypeName.value = null;
  activeLayout.value = activeLayout.value.filter(item => item.i !== 'drop-placeholder');
};

const onDrop = () => {
  const placeholderIndex = activeLayout.value.findIndex(item => item.i === 'drop-placeholder');
  if (placeholderIndex !== -1) {
    const newItem = activeLayout.value[placeholderIndex];
    newItem.i = 'new-' + Date.now().toString();

    let foundTypeId = 0;
    for (let [id, typeObj] of widgetTypeMasterMap.entries()) {
      if (typeObj.widgetTypeName === newItem.widgetTypeName) { foundTypeId = id; break; }
    }
    newItem.widgetTypeId = foundTypeId; newItem.widgetId = 0; newItem.deviceId = 0;
    newItem.widgetLabel = `New ${newItem.widgetTypeName}`;
    newItem.widgetColor = { bgHex: '#ffffff', textHex: '#334155', chartHex: '#3b82f6' };
    newItem.customChartData = {};
  }
  draggedWidgetTypeName.value = null;
  isDraggingTool.value = false;
};

const onRightClick = (event, widgetId) => {
  contextMenu.value = { show: true, x: event.clientX, y: event.clientY, widgetId: widgetId };
};

const closeContextMenu = () => { contextMenu.value.show = false; };
const closeDeleteModal = () => { deleteModal.value.close(); widgetToDelete.value = null; };
const promptDelete = () => { widgetToDelete.value = contextMenu.value.widgetId; closeContextMenu(); deleteModal.value.showModal(); };
const confirmDelete = () => { activeLayout.value = activeLayout.value.filter(item => item.i !== widgetToDelete.value); closeDeleteModal(); };

const promptConfig = () => {
  widgetToConfig.value = contextMenu.value.widgetId;
  closeContextMenu();
  const widget = activeLayout.value.find(item => item.i === widgetToConfig.value);
  if (widget) {
    configForm.value = {
      widgetLabel: widget.widgetLabel || '', deviceId: widget.deviceId || 0, widgetTypeId: widget.widgetTypeId || 0,
      widgetTypeName: widget.widgetTypeName, widgetColor: widget.widgetColor || { bgHex: '#ffffff', textHex: '#334155', chartHex: '#3b82f6' },
      layoutData: { x: widget.x, y: widget.y, w: widget.w, h: widget.h }, customChartData: widget.customChartData || {}
    };
  }
  activeTab.value = 'general';
  configModal.value.showModal();
};

const closeConfigModal = () => { configModal.value.close(); widgetToConfig.value = null; };

const saveConfig = () => {
  const widgetIndex = activeLayout.value.findIndex(item => item.i === widgetToConfig.value);
  if (widgetIndex !== -1) {
    const updatedItem = {
      ...activeLayout.value[widgetIndex], widgetLabel: configForm.value.widgetLabel, deviceId: configForm.value.deviceId,
      widgetColor: { ...configForm.value.widgetColor }, customChartData: JSON.parse(JSON.stringify(configForm.value.customChartData))
    };
    updatedItem.x = Number(configForm.value.layoutData.x); updatedItem.y = Number(configForm.value.layoutData.y);
    updatedItem.w = Number(configForm.value.layoutData.w); updatedItem.h = Number(configForm.value.layoutData.h);
    activeLayout.value.splice(widgetIndex, 1, updatedItem);
  }
  closeConfigModal();
};

onMounted(async () => {
  await setupData();
  await loadUserCanvas();
});
</script>

<style scoped>
@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-animate {
  animation: slideDown 0.25s ease-out;
}

:deep(.vgl-item__resizer) {
  width: 35px !important;
  height: 35px !important;
  right: 2px !important;
  bottom: 2px !important;
  cursor: se-resize !important;
  z-index: 10 !important;
}

.is-editable:hover+ :deep(.vgl-item__resizer)::after {
  content: '';
  position: absolute;
  right: 4px;
  bottom: 4px;
  width: 12px;
  height: 12px;
  border-right: 3px solid #3b82f6;
  border-bottom: 3px solid #3b82f6;
  border-radius: 2px;
}
</style>