<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <!-- Header -->
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide truncate pr-2">
        {{ widgetData?.widgetLabel || 'New Alert' }}
      </h3>
      
      <span v-if="hasData" class="badge font-bold shrink-0" :class="criticalCount > 0 ? 'badge-error text-white' : 'badge-success text-white'">
        {{ criticalCount }} Critical
      </span>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 w-full relative overflow-y-auto bg-white border-x border-b border-slate-200 p-2">
      
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No devices selected. Open configuration to attach monitoring sources.
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Waiting for live data...
      </div>

      <!-- Actual Alerts List -->
      <ul v-else class="flex flex-col gap-2">
        <li v-if="filteredAlerts.length === 0" class="text-center py-6 text-base-content/50 font-medium text-sm">
          No alerts to display.
        </li>
        
        <li v-for="alert in filteredAlerts" :key="alert.id" 
            class="flex items-start gap-3 p-3 rounded-lg border-l-4 shadow-sm bg-base-100/50 hover:bg-base-200/50 transition-colors"
            :style="{ borderLeftColor: getStateColor(alert.state) }">
          
          <div class="mt-0.5 shrink-0">
            <span class="flex h-3 w-3 rounded-full" :style="{ backgroundColor: getStateColor(alert.state) }"></span>
          </div>

          <div class="flex-1 min-w-0">
            <div class="flex justify-between items-start gap-2">
              <h4 class="font-bold text-sm text-base-content truncate">
                {{ alert.title }}
              </h4>
              <span class="text-xs font-semibold text-base-content/60 whitespace-nowrap">
                {{ alert.time }}
              </span>
            </div>
            
            <p v-if="!config.compactMode" class="text-xs text-base-content/70 mt-1 line-clamp-2 leading-relaxed font-mono">
              {{ alert.desc }}
            </p>
          </div>
        </li>
      </ul>

    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const liveStreamStore = useLiveStreamStore();
const deviceAlerts = ref({});

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => {
  const ids = props.widgetData?.deviceIds || [];
  return ids.some(id => liveStreamStore.liveData[id] !== undefined);
});

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    maxAlerts: customData.maxAlerts !== undefined ? customData.maxAlerts : 5,
    showResolved: customData.showResolved !== undefined ? customData.showResolved : true,
    compactMode: customData.compactMode !== undefined ? customData.compactMode : false,
    use24HourFormat: customData.use24HourFormat !== undefined ? customData.use24HourFormat : true, 
    
    critOp: customData.critOp || '>',
    critVal: customData.critVal !== undefined ? customData.critVal : 90,
    warnOp: customData.warnOp || '>',
    warnVal: customData.warnVal !== undefined ? customData.warnVal : 70,

    criticalColor: customData.criticalColor || '#ef4444', 
    warningColor: customData.warningColor || '#f59e0b',  
    resolvedColor: customData.resolvedColor || '#10b981' 
  };
});

const getStateColor = (state) => {
  if (state === 'critical') return config.value.criticalColor;
  if (state === 'warning') return config.value.warningColor;
  return config.value.resolvedColor;
};

const checkCondition = (val, op, target) => {
  if (target === undefined || target === null || target === '') return false;
  const numVal = Number(val);
  const numTarget = Number(target);
  if (isNaN(numVal) || isNaN(numTarget)) return false;

  switch(op) {
    case '>': return numVal > numTarget;
    case '>=': return numVal >= numTarget;
    case '<': return numVal < numTarget;
    case '<=': return numVal <= numTarget;
    case '==': return numVal == numTarget;
    case '!=': return numVal != numTarget;
    default: return false;
  }
};

const evaluateState = (val) => {
  if (checkCondition(val, config.value.critOp, config.value.critVal)) return 'critical';
  if (checkCondition(val, config.value.warnOp, config.value.warnVal)) return 'warning';
  return 'resolved';
};

// ⚡ THE FIX: Watch the central store and generate the alerts!
watch(() => liveStreamStore.liveData, (newData) => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0) return;

  const hasIncomingData = rawDeviceIds.some(id => newData[String(id)] !== undefined);
  if (!hasIncomingData) return;

  const newAlerts = { ...deviceAlerts.value };
  const now = Date.now();
  
  const timeStr = new Date().toLocaleTimeString(undefined, {
    hour12: !config.value.use24HourFormat,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });

  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const device = newData[id];
    
    if (device) {
      const state = evaluateState(device.value);
      
      newAlerts[id] = {
        id: id,
        state: state,
        title: device.name || `Device ${id}`,
        desc: `Latest Reading: ${device.value}`,
        time: timeStr,
        timestamp: now
      };
    }
  });

  deviceAlerts.value = newAlerts;
}, { deep: true });

// Clear alerts if assigned devices are changed
watch(
  () => props.widgetData?.deviceIds,
  (newIds, oldIds) => {
    if (sameIds(newIds, oldIds)) return; 
    deviceAlerts.value = {}; 
  },
  { deep: false }
);

function sameIds(a, b) {
  if (!a || !b || a.length !== b.length) return false;
  return a.every((id, i) => id === b[i]);
}

const criticalCount = computed(() => {
  return Object.values(deviceAlerts.value).filter(a => a.state === 'critical').length;
});

const filteredAlerts = computed(() => {
  let alerts = Object.values(deviceAlerts.value);
  
  if (!config.value.showResolved) {
    alerts = alerts.filter(a => a.state !== 'resolved');
  }
  
  const severityRank = { 'critical': 3, 'warning': 2, 'resolved': 1 };
  alerts.sort((a, b) => {
    if (severityRank[b.state] !== severityRank[a.state]) {
      return severityRank[b.state] - severityRank[a.state]; 
    }
    return b.timestamp - a.timestamp; 
  });
  
  return alerts.slice(0, config.value.maxAlerts);
});
</script>