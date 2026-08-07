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
          <span class="label-text font-semibold">Include Resolved (Green) Alerts</span>
        </label>
        
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.compactMode" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Compact Mode (Hide descriptions)</span>
        </label>
      </div>
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

// Provide sensible defaults for the Alert List
const localConfig = ref({
  maxAlerts: props.modelValue.maxAlerts !== undefined ? props.modelValue.maxAlerts : 5,
  showResolved: props.modelValue.showResolved !== undefined ? props.modelValue.showResolved : true,
  compactMode: props.modelValue.compactMode !== undefined ? props.modelValue.compactMode : false,
  criticalColor: props.modelValue.criticalColor || '#ef4444',
  warningColor: props.modelValue.warningColor || '#f59e0b',
  resolvedColor: props.modelValue.resolvedColor || '#10b981'
});

// Watch the entire object and emit ANY change automatically
watch(localConfig, (newVal) => {
  emit('update:modelValue', { ...newVal });
}, { deep: true });

// Emit the default values immediately when the modal opens
onMounted(() => {
  emit('update:modelValue', { ...localConfig.value });
});
</script>