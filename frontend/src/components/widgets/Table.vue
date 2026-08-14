<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">

    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New Table' }}
      </h3>
      <span class="badge badge-neutral badge-sm font-semibold" v-if="hasData && config.showRowCount">
        {{ displayData.length }} / {{ config.maxRows }} Rows
      </span>
    </div>

    <!-- Scrollable Table Container -->
    <div class="flex-1 w-full relative overflow-auto border-x border-b border-slate-200 bg-white flex flex-col justify-center">

      <!-- 1. No Devices Selected State -->
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No devices selected. Open configuration to add data sources.
      </div>

      <!-- 2. Loading State -->
      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Waiting for live data...
      </div>

      <!-- 3. Actual Table -->
      <table v-else class="table w-full text-left relative" :class="{ 'table-zebra': config.isStriped, 'table-sm': config.isDense }">

        <thead class="sticky top-0 z-10 shadow-sm" :style="{ backgroundColor: config.headerColor, color: config.headerTextColor }">
          <tr>
            <th v-for="(col, index) in displayColumns" :key="index" class="font-bold text-sm tracking-wide whitespace-nowrap">
              {{ col }}
            </th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="(row, index) in displayData" :key="index" class="hover:bg-slate-50 transition-colors">
            <td v-for="(col, colIndex) in displayColumns" :key="colIndex" class="border-b border-slate-100 whitespace-nowrap font-medium">
              {{ row[col] !== undefined ? row[col] : '-' }}
            </td>
          </tr>
        </tbody>

      </table>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue';

// ⚡ Import the shared Pinia store
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    isStriped: customData.isStriped !== undefined ? customData.isStriped : true,
    isDense: customData.isDense !== undefined ? customData.isDense : false,
    showRowCount: customData.showRowCount !== undefined ? customData.showRowCount : true,
    maxRows: customData.maxRows !== undefined ? customData.maxRows : 10,
    headerColor: customData.headerColor || '#f8fafc',
    headerTextColor: customData.headerTextColor || '#334155',
    use24HourFormat: customData.use24HourFormat !== undefined ? customData.use24HourFormat : true,
    showTimeColumn: customData.showTimeColumn !== undefined ? customData.showTimeColumn : true // ⚡ Read new config
  };
});

const liveStreamStore = useLiveStreamStore();
const liveTableRows = ref([]);

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => liveTableRows.value.length > 0);

// ⚡ Dynamically build the column headers
const displayColumns = computed(() => {
  const cols = [];
  
  // ⚡ Check the config to see if we should render the Time column
  if (config.value.showTimeColumn) {
    cols.push('Time');
  }

  const rawDeviceIds = props.widgetData?.deviceIds || [];
  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const device = liveStreamStore.liveData[id];
    cols.push(device ? device.name : `Device ${id}`);
  });
  
  return cols;
});

// ⚡ Watch the central store and build the table row history
watch(() => liveStreamStore.liveData, (newData) => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0) return;

  const hasIncomingData = rawDeviceIds.some(id => newData[String(id)] !== undefined);
  if (!hasIncomingData) return;

  const timeStr = new Date().toLocaleTimeString(undefined, {
    hour12: !config.value.use24HourFormat,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });

  const newRow = { 'Time': timeStr };

  // Loop through selected devices to map values to their specific column names
  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const device = newData[id];
    const colName = device ? device.name : `Device ${id}`;
    
    newRow[colName] = device ? device.value : '-';
  });

  liveTableRows.value.unshift(newRow);

  if (liveTableRows.value.length > config.value.maxRows) {
    liveTableRows.value = liveTableRows.value.slice(0, config.value.maxRows);
  }
}, { deep: true });

// Clear the table history if the user changes the assigned devices
watch(
  () => props.widgetData?.deviceIds,
  (newIds, oldIds) => {
    if (sameIds(newIds, oldIds)) return;
    liveTableRows.value = [];
  },
  { deep: false }
);

function sameIds(a, b) {
  if (!a || !b || a.length !== b.length) return false;
  return a.every((id, i) => id === b[i]);
}

const displayData = computed(() => liveTableRows.value);
</script>