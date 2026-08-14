<template>
  <div class="flex flex-col gap-5">
    
    <!-- Color Configuration -->
    <div class="grid grid-cols-1 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Default Main Text Color</span></div>
        <input type="color" v-model="localConfig.valueColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <!-- Text Formatting -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Text & Value Formatting</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Prefix</span>
          </div>
          <input type="text" v-model="localConfig.prefix"
            class="input input-bordered input-sm w-full" placeholder="e.g., $" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Unit (Suffix)</span>
          </div>
          <input type="text" v-model="localConfig.unit"
            class="input input-bordered input-sm w-full" placeholder="e.g., °C, MB/s" />
        </label>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Font Size (px)</span>
          </div>
          <input type="number" v-model.number="localConfig.fontSize"
            class="input input-bordered input-sm w-full" placeholder="56" />
        </label>

        <!-- ⚡ NEW: Changed to a textarea to support multiple lines -->
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Subtext / Description</span>
          </div>
          <textarea v-model="localConfig.subtext"
            class="textarea textarea-bordered w-full h-20" placeholder="Type here. Press Enter for new line..."></textarea>
        </label>
      </div>
    </div>

    <!-- Conditional Formatting -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">Conditional Status Display</h4>
        <input type="checkbox" v-model="localConfig.enableConditions" class="toggle toggle-primary toggle-sm" />
      </div>

      <div v-if="localConfig.enableConditions" class="flex flex-col gap-3">
        <div class="flex justify-end">
          <button @click="addCondition" class="btn btn-xs btn-primary btn-outline">
            + Add Condition
          </button>
        </div>

        <div v-for="(cond, index) in localConfig.conditions" :key="index" 
             class="flex flex-wrap items-center gap-2 p-3 bg-base-100 border border-base-300 rounded-lg">
          
          <div class="flex items-center gap-2 w-full md:w-auto">
            <span class="text-xs font-semibold whitespace-nowrap">If Value</span>
            <select v-model="cond.operator" class="select select-bordered select-xs w-16">
              <option value="==">==</option>
              <option value="!=">!=</option>
              <option value=">">&gt;</option>
              <option value=">=">&gt;=</option>
              <option value="<">&lt;</option>
              <option value="<=">&lt;=</option>
            </select>
            <input type="text" v-model="cond.compareValue" class="input input-bordered input-xs w-20" placeholder="Value" />
          </div>

          <div class="flex items-center gap-2 w-full md:w-auto flex-1">
            <span class="text-xs font-semibold whitespace-nowrap">Show:</span>
            <input type="text" v-model="cond.displayText" class="input input-bordered input-xs flex-1" placeholder="Text to display" />
            <input type="color" v-model="cond.color" class="h-6 w-8 cursor-pointer rounded border border-base-300 p-0" title="Text Color if met" />
          </div>

          <button @click="removeCondition(index)" class="btn btn-xs btn-ghost text-error hover:bg-error/10 px-2" title="Remove Condition">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <div v-if="localConfig.conditions.length === 0" class="text-xs text-base-content/50 italic text-center py-2">
          No conditions set. The raw value will be displayed.
        </div>
        <p class="text-[10px] text-base-content/60 mt-1 leading-tight">
          * Rules are evaluated from top to bottom. The first matching rule will be applied.
        </p>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';
import { Icon } from '@iconify/vue';

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);

const defaultConditions = [
  { operator: '==', compareValue: '1', displayText: 'Online', color: '#10b981' },
  { operator: '==', compareValue: '0', displayText: 'Offline', color: '#ef4444' }
];

const localConfig = ref({
  valueColor: props.modelValue.valueColor || '#10b981',
  fontSize: props.modelValue.fontSize !== undefined ? props.modelValue.fontSize : 56,
  prefix: props.modelValue.prefix || '',
  unit: props.modelValue.unit || '',
  subtext: props.modelValue.subtext || '',
  enableConditions: props.modelValue.enableConditions || false,
  conditions: props.modelValue.conditions ? JSON.parse(JSON.stringify(props.modelValue.conditions)) : defaultConditions
});

const addCondition = () => {
  localConfig.value.conditions.push({
    operator: '==',
    compareValue: '',
    displayText: 'Status Met',
    color: localConfig.value.valueColor 
  });
};

const removeCondition = (index) => {
  localConfig.value.conditions.splice(index, 1);
};

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>