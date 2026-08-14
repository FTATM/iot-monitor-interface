<template>
  <div class="flex flex-col gap-4 w-full">

    <!-- Top Toolbar with Search -->
    <div class="flex justify-between items-center w-full">
      <div class="relative w-full max-w-xs">
        <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-base-content/50">
          <Icon icon="lucide:search" class="w-4 h-4" />
        </div>
        <input type="text" v-model="globalFilter"
          class="input input-bordered w-full pl-10 focus:border-primary transition-colors shadow-sm"
          placeholder="Search all columns..." />
      </div>

      <!-- Slot for extra buttons -->
      <div>
        <slot name="toolbar-actions"></slot>
      </div>
    </div>

    <!-- The DaisyUI Table -->
    <div class="bg-base-100 rounded-box shadow-sm border border-base-200 overflow-x-auto">
      <table class="table w-full">
        <thead class="bg-base-200 text-base-content text-sm">
          <tr v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
            
            <!-- Added click handler and hover styles for sorting -->
            <th v-for="header in headerGroup.headers" :key="header.id"
              :class="[
                header.column.columnDef.meta?.headerClass,
                header.column.getCanSort() ? 'cursor-pointer select-none hover:bg-base-300/50 transition-colors' : ''
              ]"
              @click="header.column.getToggleSortingHandler()?.($event)"
            >
              <div class="flex items-center gap-1" :class="header.column.columnDef.meta?.headerClass?.includes('text-right') ? 'justify-end' : ''">
                
                <!-- NEW V9: Greatly simplified FlexRender for Headers! -->
                <FlexRender v-if="!header.isPlaceholder" :header="header" />
                  
                <!-- Sorting Icons -->
                <span v-if="header.column.getCanSort()" class="w-4 h-4 flex items-center justify-center text-base-content/50">
                  <Icon v-if="header.column.getIsSorted() === 'asc'" icon="lucide:arrow-up" class="w-3 h-3 text-primary" />
                  <Icon v-else-if="header.column.getIsSorted() === 'desc'" icon="lucide:arrow-down" class="w-3 h-3 text-primary" />
                  <Icon v-else icon="lucide:arrow-up-down" class="w-3 h-3 opacity-30" />
                </span>
              </div>
            </th>

          </tr>
        </thead>

        <tbody>
          <!-- 1. LOADING STATE -->
          <tr v-if="isLoading">
            <td :colspan="table.getAllColumns().length" class="text-center py-12">
              <span class="loading loading-spinner loading-lg text-primary"></span>
            </td>
          </tr>

          <!-- Empty State -->
          <tr v-if="table.getRowModel().rows.length === 0">
            <td :colspan="table.getAllColumns().length" class="text-center py-12 text-base-content/50">
              No matching records found.
            </td>
          </tr>

          <!-- Data Rows -->
          <tr v-for="row in table.getRowModel().rows" :key="row.id" class="hover:bg-base-200/30 transition-colors">
            <td v-for="cell in row.getAllCells()" :key="cell.id" :class="cell.column.columnDef.meta?.cellClass">
              
              <!-- MAGIC SLOT -->
              <slot :name="'cell-' + cell.column.id" :row="row.original" :value="cell.getValue()">
                <!-- NEW V9: Greatly simplified FlexRender for Cells! -->
                <FlexRender :cell="cell" />
              </slot>
              
            </td>
          </tr>
        </tbody>
      </table>
    </div>

  </div>
</template>

<script setup>
import { ref } from 'vue';
import { Icon } from '@iconify/vue';
import {
  useTable,
  tableFeatures,
  globalFilteringFeature,
  createFilteredRowModel,
  filterFn_includesString,
  rowSortingFeature,       // <-- Updated from docs!
  createSortedRowModel,
  sortFn_alphanumeric,     // <-- Updated from docs!
  sortFn_text,             // <-- Updated from docs!
  FlexRender
} from '@tanstack/vue-table';

const props = defineProps({
  data: {
    type: Array,
    required: true
  },
  columns: {
    type: Array,
    required: true
  },
  initialSorting: {
    type: Array,
    default: () => [] 
  },
  isLoading: {
    type: Boolean,
    default: false
  }
});

// State definitions
const globalFilter = ref('');
const sorting = ref(props.initialSorting);

// Setup V9 Features
const features = tableFeatures({
  globalFilteringFeature,
  rowSortingFeature, // Updated
  filteredRowModel: createFilteredRowModel(),
  sortedRowModel: createSortedRowModel(),
  filterFns: {
    includesString: filterFn_includesString
  },
  sortFns: { // Explicitly defining the default sort functions from docs
    alphanumeric: sortFn_alphanumeric,
    text: sortFn_text,
  }
});

// Initialize the TanStack Table instance
const table = useTable({
  features,
  get data() { return props.data; },
  get columns() { return props.columns; },
  state: {
    get globalFilter() { return globalFilter.value; },
    get sorting() { return sorting.value; }
  },
  onGlobalFilterChange: (updaterOrValue) => {
    globalFilter.value = typeof updaterOrValue === 'function'
      ? updaterOrValue(globalFilter.value)
      : updaterOrValue;
  },
  onSortingChange: (updaterOrValue) => {
    sorting.value = typeof updaterOrValue === 'function'
      ? updaterOrValue(sorting.value)
      : updaterOrValue;
  },
  globalFilterFn: 'includesString',
  enableSortingRemoval: false
});
</script>