<!-- src/components/SearchableDropdown.vue -->
<template>
  <div class="relative w-full">
    <!-- Search Input -->
    <input 
      type="text" 
      v-model="searchQuery" 
      @focus="isOpen = true" 
      @blur="handleBlur"
      :placeholder="placeholder" 
      :class="['input input-bordered w-full', { 'input-error': error }]" 
    />
    
    <!-- Dropdown List -->
    <ul 
      v-show="isOpen" 
      class="absolute z-50 w-full p-2 mt-1 shadow-xl bg-base-100 rounded-box max-h-48 overflow-y-auto border border-base-200"
    >
      <li v-for="option in filteredOptions" :key="option[valueKey]">
        <a 
          class="block px-4 py-2 hover:bg-base-200 cursor-pointer rounded-lg font-medium" 
          @mousedown.prevent="selectOption(option)"
        >
          {{ option[labelKey] }}
        </a>
      </li>
      <li v-if="filteredOptions.length === 0" class="px-4 py-2 text-base-content/50 text-sm">
        No results found for "{{ searchQuery }}"
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: null
  },
  options: {
    type: Array,
    default: () => []
  },
  // Allows the component to adapt to any database schema
  labelKey: {
    type: String,
    default: 'name' 
  },
  valueKey: {
    type: String,
    default: 'id'
  },
  placeholder: {
    type: String,
    default: 'Search...'
  },
  error: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:modelValue', 'blur']);

const searchQuery = ref('');
const isOpen = ref(false);

// Filter options based on what the user types
const filteredOptions = computed(() => {
  if (!searchQuery.value) return props.options;
  return props.options.filter(opt => 
    String(opt[props.labelKey]).toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});

const selectOption = (option) => {
  searchQuery.value = option[props.labelKey];
  isOpen.value = false;
  // Update the parent's v-model
  emit('update:modelValue', option[props.valueKey]);
  emit('blur'); // Let the parent know to trigger Vuelidate
};

// Handle clicking outside the input
const handleBlur = () => {
  isOpen.value = false;
  
  // If they typed a partial word and clicked away, reset the text 
  // to the currently selected modelValue, or clear it if nothing is selected.
  const selected = props.options.find(opt => opt[props.valueKey] === props.modelValue);
  searchQuery.value = selected ? selected[props.labelKey] : '';
  
  emit('blur');
};

// Watch for external data changes (like clicking "Edit" on a table row, or API loading)
watch(
  () => [props.modelValue, props.options],
  () => {
    const selected = props.options.find(opt => opt[props.valueKey] === props.modelValue);
    searchQuery.value = selected ? selected[props.labelKey] : '';
  },
  { immediate: true, deep: true }
);
</script>