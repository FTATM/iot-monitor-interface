<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <!-- Header -->
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New Alert' }}
      </h3>
      <span class="badge font-bold" :class="filteredAlerts.length > 0 ? 'badge-error text-white' : 'badge-success text-white'">
        {{ filteredAlerts.filter(a => a.state === 'critical').length }} Critical
      </span>
    </div>

    <!-- Scrollable Alert List -->
    <div class="flex-1 w-full relative overflow-y-auto bg-white border-x border-b border-slate-200 p-2">
      <ul class="flex flex-col gap-2">
        <li v-if="filteredAlerts.length === 0" class="text-center py-6 text-base-content/50 font-medium text-sm">
          No alerts to display.
        </li>
        
        <li v-for="alert in filteredAlerts" :key="alert.id" 
            class="flex items-start gap-3 p-3 rounded-lg border-l-4 shadow-sm bg-base-100/50 hover:bg-base-200/50 transition-colors"
            :style="{ borderLeftColor: getStateColor(alert.state) }">
          
          <!-- Icon / Indicator -->
          <div class="mt-0.5 shrink-0">
            <span class="flex h-3 w-3 rounded-full" :style="{ backgroundColor: getStateColor(alert.state) }"></span>
          </div>

          <!-- Alert Content -->
          <div class="flex-1 min-w-0">
            <div class="flex justify-between items-start gap-2">
              <h4 class="font-bold text-sm text-base-content truncate">
                {{ alert.title }}
              </h4>
              <span class="text-xs font-semibold text-base-content/60 whitespace-nowrap">
                {{ alert.time }}
              </span>
            </div>
            
            <!-- Description (Hidden in Compact Mode) -->
            <p v-if="!config.compactMode" class="text-xs text-base-content/70 mt-1 line-clamp-2 leading-relaxed">
              {{ alert.desc }}
            </p>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

// Extract configurations with safe fallbacks
const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    maxAlerts: customData.maxAlerts !== undefined ? customData.maxAlerts : 5,
    showResolved: customData.showResolved !== undefined ? customData.showResolved : true,
    compactMode: customData.compactMode !== undefined ? customData.compactMode : false,
    criticalColor: customData.criticalColor || '#ef4444', // Red
    warningColor: customData.warningColor || '#f59e0b',  // Amber
    resolvedColor: customData.resolvedColor || '#10b981' // Emerald
  };
});

// Helper to map the state to the configured colors
const getStateColor = (state) => {
  if (state === 'critical') return config.value.criticalColor;
  if (state === 'warning') return config.value.warningColor;
  return config.value.resolvedColor;
};

// Dummy Data mapped to common backend/system scenarios
const allDummyAlerts = [
  { id: 1, state: 'critical', title: 'High CPU Usage', desc: 'Server CPU exceeded 95% for 5 minutes.', time: '2m ago' },
  { id: 2, state: 'warning', title: 'S3 Storage Migration Delay', desc: 'Wasabi S3 migration tool encountering high latency during bulk SQL updates.', time: '15m ago' },
  { id: 3, state: 'resolved', title: 'Database Deadlock', desc: 'Transaction deadlock on financial records resolved automatically.', time: '1h ago' },
  { id: 4, state: 'critical', title: 'Authentication Service', desc: 'Login session token validation failing.', time: '2h ago' },
  { id: 5, state: 'warning', title: 'Memory Usage High', desc: 'Local node_modules process utilizing excessive RAM.', time: '3h ago' },
  { id: 6, state: 'resolved', title: 'Network Spike', desc: 'Ingress traffic normalized.', time: '5h ago' }
];

// Filter and limit the alerts based on widget configuration
const filteredAlerts = computed(() => {
  let alerts = allDummyAlerts;
  
  if (!config.value.showResolved) {
    alerts = alerts.filter(a => a.state !== 'resolved');
  }
  
  return alerts.slice(0, config.value.maxAlerts);
});
</script>