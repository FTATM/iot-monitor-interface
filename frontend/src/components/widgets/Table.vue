<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New Table' }}
      </h3>
      <span class="badge badge-neutral badge-sm" v-if="config.showRowCount">
        {{ dummyData.length }} Rows
      </span>
    </div>

    <!-- Scrollable Table Container -->
    <div class="flex-1 w-full relative overflow-auto border-x border-b border-slate-200 bg-white">
      <table class="table w-full text-left" 
             :class="{ 'table-zebra': config.isStriped, 'table-sm': config.isDense }">
        
        <!-- Table Header -->
        <thead class="sticky top-0 z-10 shadow-sm" :style="{ backgroundColor: config.headerColor, color: config.headerTextColor }">
          <tr>
            <th v-for="(col, index) in parsedColumns" :key="index" class="font-bold text-sm tracking-wide">
              {{ col }}
            </th>
          </tr>
        </thead>
        
        <!-- Table Body (Dummy Data for Editor) -->
        <tbody>
          <tr v-for="row in dummyData" :key="row.id" class="hover:bg-slate-50 transition-colors">
            <!-- Dynamically match row data to the defined columns -->
            <td v-for="(col, colIndex) in parsedColumns" :key="colIndex" class="border-b border-slate-100">
              <!-- Render a status badge if the column is named 'Status' -->
              <span v-if="col.toLowerCase() === 'status'" 
                    class="badge badge-sm font-semibold"
                    :class="row[col] === 'Completed' ? 'badge-success text-white' : 'badge-warning'">
                {{ row[col] || 'Pending' }}
              </span>
              <!-- Render normal text otherwise -->
              <span v-else>
                {{ row[col] || '-' }}
              </span>
            </td>
          </tr>
        </tbody>

      </table>
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
    columns: customData.columns || 'ID, Date, Description, Status, Amount',
    isStriped: customData.isStriped !== undefined ? customData.isStriped : true,
    isDense: customData.isDense !== undefined ? customData.isDense : false,
    showRowCount: customData.showRowCount !== undefined ? customData.showRowCount : true,
    headerColor: customData.headerColor || '#f8fafc', // Slate 50
    headerTextColor: customData.headerTextColor || '#334155' // Slate 700
  };
});

// Convert the comma-separated string into an array of column names
const parsedColumns = computed(() => {
  return config.value.columns.split(',').map(c => c.trim()).filter(c => c.length > 0);
});

// Generate sensible dummy data that matches whatever columns the user typed
const dummyData = computed(() => {
  const data = [];
  const statuses = ['Completed', 'Pending', 'Completed', 'Failed'];
  
  for (let i = 1; i <= 5; i++) {
    let row = { id: i };
    parsedColumns.value.forEach(col => {
      const colLower = col.toLowerCase();
      if (colLower === 'id') row[col] = `REQ-${1000 + i}`;
      else if (colLower === 'date') row[col] = `2026-05-0${i}`;
      else if (colLower === 'status') row[col] = statuses[i % 4];
      else if (colLower === 'amount') row[col] = `$${(Math.random() * 1000).toFixed(2)}`;
      else if (colLower === 'description') row[col] = `Migration Task #${i}`;
      else row[col] = `Data ${i}`;
    });
    data.push(row);
  }
  return data;
});
</script>