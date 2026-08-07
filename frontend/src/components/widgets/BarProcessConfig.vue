<template>
  <div class="flex flex-col gap-5">
    
    <!-- Colors -->
    <div class="grid grid-cols-2 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Progress Color</span></div>
        <input type="color" v-model="localConfig.progressColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Background Track Color</span></div>
        <input type="color" v-model="localConfig.trackColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <!-- Target & Label -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Data Configuration</h4>
      
      <div class="grid grid-cols-1 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Maximum Goal (100%)</span>
          </div>
          <input type="number" v-model.number="localConfig.maxValue"
            class="input input-bordered input-sm w-full" placeholder="100" />
        </label>
      </div>

      <label class="cursor-pointer label justify-start gap-4">
        <input type="checkbox" v-model="localConfig.showTextLabel" class="toggle toggle-primary" />
        <span class="label-text font-semibold">Show Progress Numbers in Header</span>
      </label>
    </div>

    <!-- Sizing & Style -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Bar Styling</h4>
      
      <div class="grid grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Thickness (px)</span>
          </div>
          <input type="number" v-model.number="localConfig.barThickness"
            class="input input-bordered input-sm w-full" placeholder="24" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Corner Radius (px)</span>
          </div>
          <input type="number" v-model.number="localConfig.borderRadius"
            class="input input-bordered input-sm w-full" placeholder="12" />
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

// Provide sensible defaults for a Progress Bar
const localConfig = ref({
  progressColor: props.modelValue.progressColor || '#10b981',
  trackColor: props.modelValue.trackColor || '#e2e8f0',
  maxValue: props.modelValue.maxValue !== undefined ? props.modelValue.maxValue : 100,
  barThickness: props.modelValue.barThickness !== undefined ? props.modelValue.barThickness : 24,
  borderRadius: props.modelValue.borderRadius !== undefined ? props.modelValue.borderRadius : 12,
  showTextLabel: props.modelValue.showTextLabel !== undefined ? props.modelValue.showTextLabel : true
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