<template>
  <div class="flex flex-col gap-5">
    
    <!-- Gauge Scale Configuration -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Scale Settings</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-3">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">Minimum Value</span></div>
          <input type="number" v-model.number="localConfig.min"
            class="input input-bordered input-sm w-full transition-colors"
            :class="{ 'input-error text-error font-bold': hasMinMismatch || localConfig.min >= localConfig.max }"
            placeholder="0" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">Maximum Value</span></div>
          <input type="number" v-model.number="localConfig.max"
            class="input input-bordered input-sm w-full transition-colors" 
            :class="{ 'input-error text-error font-bold': hasMaxMismatch || localConfig.min >= localConfig.max }"
            placeholder="100" />
        </label>
      </div>

      <label class="form-control w-full">
        <div class="label pb-1">
          <span class="label-text font-semibold">Unit Label</span>
        </div>
        <input type="text" v-model="localConfig.unit"
          class="input input-bordered input-sm w-full" placeholder="%" />
      </label>
    </div>

    <!-- Grade Text Display -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Grade Labels</h4>
      
      <div class="flex flex-col gap-3">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showGradeText" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Show Grade Text</span>
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">Text Size</span></div>
          <input type="number" v-model.number="localConfig.gradeTextSize"
            class="input input-bordered input-sm w-full" :disabled="!localConfig.showGradeText" placeholder="14" />
        </label>
      </div>
    </div>

    <!-- Dynamic Grade Zoning -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">Grade Zones</h4>
        <button @click="addGrade" class="btn btn-primary btn-xs">
          + Add Zone
        </button>
      </div>

      <!-- Strict Warning 1: Highest limit doesn't match Max -->
      <div v-if="hasMaxMismatch" class="alert alert-error shadow-sm mb-4 py-2 px-3">
        <Icon icon="lucide:circle-x" class="w-5 h-5"/>
        <span class="text-sm"><strong>Error:</strong> The highest grade limit must exactly match the Maximum Value ({{ localConfig.max }}).</span>
      </div>

      <!-- Strict Warning 2: A limit drops below or equals Min -->
      <div v-if="hasMinMismatch" class="alert alert-error shadow-sm mb-4 py-2 px-3">
        <Icon icon="lucide:circle-x" class="w-5 h-5"/>
        <span class="text-sm"><strong>Error:</strong> All grade limits must be strictly greater than the Minimum Value ({{ localConfig.min }}).</span>
      </div>

      <!-- Strict Warning 3: Invalid Min/Max span -->
      <div v-if="localConfig.min >= localConfig.max" class="alert alert-error shadow-sm mb-4 py-2 px-3">
        <Icon icon="lucide:circle-x" class="w-5 h-5"/>
        <span class="text-sm"><strong>Error:</strong> Minimum Value must be strictly less than the Maximum Value.</span>
      </div>
      
      <div class="flex flex-col gap-3">
        <div v-for="(grade, index) in localConfig.grades" :key="index" class="flex flex-wrap md:flex-nowrap items-center gap-2 bg-base-100 p-2 rounded border border-base-200">
          
          <input type="color" v-model="grade.color" class="h-8 w-10 cursor-pointer rounded border border-base-300 p-0 shrink-0" />
          
          <input type="text" v-model="grade.name" class="input input-bordered input-sm w-full md:w-1/2" placeholder="Grade Name" />
          
          <div class="flex items-center gap-2 w-full md:w-auto">
            <span class="text-xs font-semibold whitespace-nowrap">Up to:</span>
            <!-- Input turns red if it falls outside the min/max span completely -->
            <input type="number" v-model.number="grade.limit" 
              class="input input-bordered input-sm w-full md:w-24 transition-colors" 
              :class="{ 'input-error text-error font-bold': grade.limit <= localConfig.min || grade.limit > localConfig.max }" />
          </div>

          <button @click="removeGrade(index)" class="btn btn-ghost btn-sm btn-square text-error shrink-0" title="Remove Grade">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div v-if="localConfig.grades.length === 0" class="text-xs text-base-content/50 italic text-center py-2">
          No grade zones configured.
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';
import { Icon } from '@iconify/vue'; 

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);

const defaultGrades = [
  { name: 'Grade D', limit: 25, color: '#FF6E76' },
  { name: 'Grade C', limit: 50, color: '#FDDD60' },
  { name: 'Grade B', limit: 75, color: '#58D9F9' },
  { name: 'Grade A', limit: 100, color: '#7CFFB2' }
];

const localConfig = ref({
  min: props.modelValue.min !== undefined ? props.modelValue.min : 0,
  max: props.modelValue.max !== undefined ? props.modelValue.max : 100,
  unit: props.modelValue.unit || '',
  showGradeText: props.modelValue.showGradeText !== undefined ? props.modelValue.showGradeText : true,
  gradeTextSize: props.modelValue.gradeTextSize !== undefined ? props.modelValue.gradeTextSize : 14,
  grades: props.modelValue.grades ? JSON.parse(JSON.stringify(props.modelValue.grades)) : defaultGrades
});

// Determine the absolute highest limit dynamically
const highestLimit = computed(() => {
  if (localConfig.value.grades.length === 0) return null;
  return Math.max(...localConfig.value.grades.map(g => g.limit));
});

// Enforce that the highest grade must perfectly match the max value
const hasMaxMismatch = computed(() => {
  return localConfig.value.grades.length > 0 && highestLimit.value !== localConfig.value.max;
});

// Enforce that no grade drops to or below the min value
const hasMinMismatch = computed(() => {
  return localConfig.value.grades.some(grade => grade.limit <= localConfig.value.min);
});

// Lock saving if any of the boundary conditions fail
const isInvalid = computed(() => {
  return (
    hasMaxMismatch.value || 
    hasMinMismatch.value || 
    localConfig.value.min >= localConfig.value.max ||
    localConfig.value.grades.length === 0
  );
});

const addGrade = () => {
  localConfig.value.grades.push({ 
    name: 'New Grade', 
    limit: localConfig.value.max, 
    color: '#3b82f6' 
  });
};

const removeGrade = (index) => {
  localConfig.value.grades.splice(index, 1);
};

watch(localConfig, (newVal) => {
  emit('update:modelValue', { ...newVal, _isInvalid: isInvalid.value });
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', { ...localConfig.value, _isInvalid: isInvalid.value });
});
</script>