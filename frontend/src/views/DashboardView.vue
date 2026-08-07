<template>
  <div class="flex h-screen w-full bg-base-200 relative overflow-hidden">
    <div class="flex-1 w-full p-6 overflow-y-auto">

      <!-- Top Control Bar (View Only) -->
      <div
        class="flex flex-wrap md:flex-nowrap justify-between items-center gap-4 mb-6 bg-base-100 p-4 rounded-box shadow-sm border border-base-200">
        <div class="flex items-center gap-3 font-medium text-base-content overflow-x-auto">
          <label for="canvas-select" class="whitespace-nowrap">Active Canvas:</label>
          <select id="canvas-select" v-model="activeCanvasId" @change="loadCurrentCanvas()"
            class="select select-bordered select-sm min-w-[200px]">
            <option v-for="[id, canvas] in allUserCanvasesMap" :key="id" :value="id">
              {{ canvas.name }}
            </option>
          </select>
          <span
            class="badge badge-neutral badge-sm font-bold uppercase tracking-wider py-3 px-3 whitespace-nowrap shrink-0">
            View Only
          </span>
        </div>
      </div>

      <!-- Grid Layout (Non-Draggable, Non-Resizable) -->
      <div class="flex-1 min-h-[500px] w-full">
        <GridLayout v-model:layout="activeLayout" :col-num="12" :row-height="30" :is-draggable="false"
          :is-resizable="false">
          <GridItem v-for="item in activeLayout" :key="item.i" :x="item.x" :y="item.y" :w="item.w" :h="item.h"
            :i="item.i">
            <div
              class="w-full h-full bg-base-100 border border-base-200 rounded-box shadow-sm overflow-hidden cursor-default widget-container flex flex-col">
              <component :is="widgetMap[item.widgetTypeName]" :widget-data="item" />
            </div>
          </GridItem>
        </GridLayout>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { GridLayout, GridItem } from 'grid-layout-plus';
import { useFetch } from '@/composables/useFetch';
import { toast } from 'vue3-toastify';

import BarChart from '@/components/widgets/BarChart.vue';
import BulletChart from '@/components/widgets/BulletChart.vue';
import GaugeChart from '@/components/widgets/GaugeChart.vue';
import LineChart from '@/components/widgets/LineChart.vue';
import PieChart from '@/components/widgets/PieChart.vue'
import ScatterChart from '@/components/widgets/ScatterChart.vue'
import BarProcess from '@/components/widgets/BarProcess.vue'
import Status from '@/components/widgets/Status.vue'
import Table from '@/components/widgets/Table.vue'
import Alert from '@/components/widgets/Alert.vue'
import Text from '@/components/widgets/Text.vue'

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
  'Text':Text,
};

const allUserCanvasesMap = ref(new Map());
const activeCanvasId = ref(null);
const activeLayout = ref([]);
const widgetTypeMasterMap = new Map();

const { data: widgetTypeMasterData, error: widgetTypeMasterError, execute: widgetTypeMasterApi } = useFetch();
const { data: userAllCanvasFetchData, error: userAllCanvasFetchError, execute: userAllCanvasFetchApi } = useFetch();

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

onMounted(async () => {
  await setupData();
  await loadUserCanvas();
});
</script>