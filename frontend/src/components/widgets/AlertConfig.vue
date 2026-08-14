<template>
  <div class="flex flex-col gap-5">
    
    <!-- Display Toggles -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">List Display Options</h4>
      
      <div class="grid grid-cols-1 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Maximum Alerts to Show</span>
          </div>
          <input type="number" v-model.number="localConfig.maxAlerts"
            class="input input-bordered input-sm w-full" placeholder="5" />
        </label>
      </div>

      <div class="flex flex-col gap-3">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showResolved" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Include Normal/Resolved (Green) Status</span>
        </label>
        
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.compactMode" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Compact Mode (Hide descriptions)</span>
        </label>
        
        <!-- ⚡ NEW: 12/24 Hour Time Format Toggle -->
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.use24HourFormat" class="toggle toggle-secondary toggle-sm" />
          <span class="label-text font-semibold">Use 24-Hour Time Format</span>
        </label>
      </div>
    </div>

    <!-- Live Alert Threshold Logic -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Alert Thresholds</h4>
      
      <!-- Critical Rule -->
      <div class="flex items-center flex-wrap gap-2 mb-3">
        <span class="badge badge-error text-white font-bold w-20 shrink-0">Critical</span>
        <span class="text-sm font-semibold">if value is</span>
        <select v-model="localConfig.critOp" class="select select-bordered select-sm w-16">
          <option value=">">&gt;</option>
          <option value=">=">&gt;=</option>
          <option value="<">&lt;</option>
          <option value="<=">&lt;=</option>
          <option value="==">==</option>
          <option value="!=">!=</option>
        </select>
        <input type="number" v-model.number="localConfig.critVal" class="input input-bordered input-sm w-24" placeholder="90" />
      </div>

      <!-- Warning Rule -->
      <div class="flex items-center flex-wrap gap-2">
        <span class="badge badge-warning text-white font-bold w-20 shrink-0">Warning</span>
        <span class="text-sm font-semibold">if value is</span>
        <select v-model="localConfig.warnOp" class="select select-bordered select-sm w-16">
          <option value=">">&gt;</option>
          <option value=">=">&gt;=</option>
          <option value="<">&lt;</option>
          <option value="<=">&lt;=</option>
          <option value="==">==</option>
          <option value="!=">!=</option>
        </select>
        <input type="number" v-model.number="localConfig.warnVal" class="input input-bordered input-sm w-24" placeholder="70" />
      </div>
      
      <p class="text-[10px] text-base-content/50 mt-3 leading-tight">
        * Rules evaluate top to bottom. If neither match, the device is considered Normal (Resolved).
      </p>
    </div>

    <!-- Severity Colors -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Severity Colors</h4>
      
      <div class="grid grid-cols-3 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-bold text-xs">Critical</span></div>
          <input type="color" v-model="localConfig.criticalColor"
            class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-bold text-xs">Warning</span></div>
          <input type="color" v-model="localConfig.warningColor"
            class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-bold text-xs">Resolved</span></div>
          <input type="color" v-model="localConfig.resolvedColor"
            class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
        </label>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);

const localConfig = ref({
  maxAlerts: props.modelValue.maxAlerts !== undefined ? props.modelValue.maxAlerts : 5,
  showResolved: props.modelValue.showResolved !== undefined ? props.modelValue.showResolved : true,
  compactMode: props.modelValue.compactMode !== undefined ? props.modelValue.compactMode : false,
  use24HourFormat: props.modelValue.use24HourFormat !== undefined ? props.modelValue.use24HourFormat : true, // ⚡ Default to 24-hour
  
  critOp: props.modelValue.critOp || '>',
  critVal: props.modelValue.critVal !== undefined ? props.modelValue.critVal : 90,
  warnOp: props.modelValue.warnOp || '>',
  warnVal: props.modelValue.warnVal !== undefined ? props.modelValue.warnVal : 70,

  criticalColor: props.modelValue.criticalColor || '#ef4444',
  warningColor: props.modelValue.warningColor || '#f59e0b',
  resolvedColor: props.modelValue.resolvedColor || '#10b981'
});

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>