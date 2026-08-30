<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-t-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel || $t('tableWidget.newTable') }}
      </h3>
      <span class="badge badge-neutral badge-sm font-semibold" v-if="hasData && config.showRowCount">
        {{ displayData.length }} / {{ config.maxRows }} {{ $t('common.rows') }}
      </span>
    </div>

    <div class="flex-1 w-full relative overflow-auto border-x border-b border-base-200/50 bg-base-100/30 backdrop-blur-sm rounded-b-lg flex flex-col justify-center">

      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevicesConfig') }}
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.waitingData') }}
      </div>

      <table v-else class="table w-full text-left relative" :class="{ 'table-zebra': config.isStriped, 'table-sm': config.isDense }">
        <thead class="sticky top-0 z-10 shadow-sm" :style="{ backgroundColor: config.headerColor, color: config.headerTextColor }">
          <tr>
            <th v-for="(col, index) in displayColumns" :key="index" class="font-bold text-sm tracking-wide whitespace-nowrap">
              {{ col }}
            </th>
          </tr>
        </thead>
        <tbody :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
          <tr v-for="(row, index) in displayData" :key="index" class="hover:bg-base-200/30 transition-colors">
            <td v-for="(col, colIndex) in displayColumns" :key="colIndex" class="border-b border-base-200/50 whitespace-nowrap font-medium">
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
import { useI18n } from 'vue-i18n';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1; 
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
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
    showTimeColumn: customData.showTimeColumn !== undefined ? customData.showTimeColumn : true 
  };
});

const liveStreamStore = useLiveStreamStore();
const liveTableRows = ref([]);

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => liveTableRows.value.length > 0);

const displayColumns = computed(() => {
  const cols = [];
  if (config.value.showTimeColumn) cols.push(t('common.time'));

  const rawDeviceIds = props.widgetData?.deviceIds || [];
  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const device = liveStreamStore.liveData[id];
    cols.push(device ? device.name : `${t('common.device')} ${id}`);
  });
  return cols;
});

watch(() => liveStreamStore.liveData, (newData) => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0) return;

  const hasIncomingData = rawDeviceIds.some(id => newData[String(id)] !== undefined);
  if (!hasIncomingData) return;

  const timeStr = new Date().toLocaleTimeString(undefined, {
    hour12: !config.value.use24HourFormat, hour: '2-digit', minute: '2-digit', second: '2-digit'
  });

  const newRow = {};
  if (config.value.showTimeColumn) {
    newRow[t('common.time')] = timeStr;
  }

  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const device = newData[id];
    const colName = device ? device.name : `${t('common.device')} ${id}`;
    newRow[colName] = device ? device.value : '-';
  });

  liveTableRows.value.unshift(newRow);
  if (liveTableRows.value.length > config.value.maxRows) {
    liveTableRows.value = liveTableRows.value.slice(0, config.value.maxRows);
  }
}, { deep: true });

watch(() => props.widgetData?.deviceIds, (newIds, oldIds) => {
  if (sameIds(newIds, oldIds)) return;
  liveTableRows.value = [];
}, { deep: false });

function sameIds(a, b) {
  if (!a || !b || a.length !== b.length) return false;
  return a.every((id, i) => id === b[i]);
}

const displayData = computed(() => liveTableRows.value);
</script>