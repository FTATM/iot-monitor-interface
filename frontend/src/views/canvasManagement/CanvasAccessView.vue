<template>
  <div v-if="hasPermission('Canvas Access', 'Display')" class="p-4 sm:p-6 w-full mx-auto">

    <!-- Header -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">

      <div class="flex items-center gap-4">
        <!-- Iconify User Icon -->
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:shield-check" class="w-7 h-7" />
        </div>

        <!-- Header Text -->
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">Canvas Access Management</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">Assign which roles are permitted to view
            specific canvases.</p>
        </div>
      </div>
    </div>

    <!-- 3-Column Table -->
    <TableData :data="roleTable" :columns="tableColumns" :initial-sorting="[{ id: 'roleName', desc: false }]"
      :is-loading="isRolesLoading">
      <!-- Custom Cell Slot: 'canvasId' -->
      <template #cell-roleName="{ value }">
        <span class="">{{ value }}</span>
      </template>

      <template #cell-assignedCanvases="{ row }">
        <div class="flex flex-wrap gap-2">
          <span v-for="canvasId in getAssignedCanvases(row.roleId)" :key="canvasId"
            class="badge badge-primary badge-outline badge-sm font-medium">
            {{ getCanvasName(canvasId) }}
          </span>

          <span v-if="getAssignedCanvases(row.roleId).length === 0" class="text-base-content/40 text-sm italic">
            No access granted
          </span>
        </div>
      </template>

      <!-- Custom Cell Slot: 'actions' -->
      <template #cell-actions="{ row }">
        <div class="flex justify-end gap-2">
          <button @click="openEditModal(row)" class="btn btn-sm btn-primary">
            <Icon icon="lucide:pencil" class="w-5 h-5" />
          </button>
        </div>
      </template>

    </TableData>

    <!-- Edit Access Modal -->
    <dialog ref="editModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-base-content mb-4 flex items-center gap-2">
          Manage Access for: <span class="text-primary">{{ editingRole?.roleName }}</span>
        </h3>

        <!-- Multi-Select Checkbox List -->
        <div
          class="form-control w-full max-h-[300px] overflow-y-auto bg-base-200/50 p-4 rounded-box border border-base-200 gap-1">
          <label v-for="canvas in canvasList" :key="canvas.canvasId"
            class="label cursor-pointer justify-start gap-4 hover:bg-base-200 p-2 rounded-lg transition-colors">
            <!-- v-model binds directly to the selectedCanvases array -->
            <input type="checkbox" :value="canvas.canvasId" v-model="selectedCanvases"
              class="checkbox checkbox-primary checkbox-sm" />
            <span class="label-text font-medium text-base">{{ canvas.canvasName }}</span>
          </label>

          <div v-if="canvasList.length === 0" class="text-center py-4 text-sm text-base-content/50">
            No canvases available in the system.
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="modal-action mt-6">
          <button type="button" @click="closeEditModal" class="btn btn-ghost" :disabled="isSaving">Cancel</button>
          <button type="button" @click="saveAccess" class="btn btn-primary text-white" :disabled="isSaving">
            <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
            Save Changes
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeEditModal">close</button>
      </form>
    </dialog>

  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import TableData from '@/components/TableData.vue';

// --- COMPOSABLES ---
const { data: roleData, isLoading: isRolesLoading, error: roleError, execute: fetchRolesApi } = useFetch();
const { data: canvasData, error: canvasError, execute: fetchCanvasesApi } = useFetch();
const { data: mappingData, error: mappingError, execute: fetchMappingApi } = useFetch();
const { res: updateRes, isLoading: isSaving, error: updateError, execute: updateMappingApi } = useMutation();

// --- STORE ---
const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

// --- STATE ---
const roleTable = ref([]);
const canvasList = ref([]);
const roleCanvasMap = ref(new Map()); // Map: roleId -> Array of canvasIds

const editModal = ref(null);
const editingRole = ref(null);
const selectedCanvases = ref([]); // Holds the checkbox selections in the modal

const tableColumns = [
  {
    header: 'Role Name',
    accessorKey: 'roleName'
  },
  {
    header: 'Assigned Canvases',
    id: 'assignedCanvases',
    enableSorting: false,
  },
  {
    header: 'Actions',
    id: 'actions', // Use 'id' when there isn't a direct data key
    enableSorting: false,
    meta: { headerClass: 'text-right', cellClass: 'text-right' }
  }
];

// --- DATA LOADING ---
const setupData = async () => {
  await fetchRolesApi('/role/getall');
  if (!roleError.value && roleData.value) {
    roleTable.value = roleData.value.data;
  } else {
    toast.error("Failed to load roles");
  }

  await fetchCanvasesApi('/canvas/getall');
  if (!canvasError.value && canvasData.value) {
    canvasList.value = canvasData.value.data;
  } else {
    toast.error("Failed to load canvases");
  }

  await fetchMappingApi('/canvas/getallcanvasroledetail');
  if (!mappingError.value && mappingData.value) {
    roleCanvasMap.value.clear()
    for (let detail of mappingData.value.data) {
      roleCanvasMap.value.set(detail.roleId, detail.canvasIds || []);
    }
  } else if (mappingError.value) {
    toast.error("Failed to load canvas access mapping");
  }
};

// --- HELPER METHODS ---
const getCanvasName = (id) => {
  const found = canvasList.value.find(c => c.canvasId === id);
  return found ? found.canvasName : 'Unknown Canvas';
};

const getAssignedCanvases = (roleId) => {
  return roleCanvasMap.value.get(roleId) || [];
};

// --- MODAL ACTIONS ---
const openEditModal = (role) => {
  editingRole.value = role;

  // Clone the array so we don't accidentally mutate the table before saving
  const currentAccess = roleCanvasMap.value.get(role.roleId) || [];
  selectedCanvases.value = [...currentAccess];

  editModal.value.showModal();
};

const closeEditModal = () => {
  editModal.value.close();
  editingRole.value = null;
  selectedCanvases.value = [];
};

const saveAccess = async () => {
  if (!editingRole.value) return;

  // The payload sends the RoleId and an array of all selected CanvasIds
  const payload = {
    roleId: editingRole.value.roleId,
    canvasIds: selectedCanvases.value
  };

  // NOTE: Ensure this matches the route where you handle the update in your Go backend
  await updateMappingApi('/canvas/upsertcanvasrole', payload, 'POST');

  if (!updateError.value && updateRes.value?.ok) {
    // Update the frontend map to instantly reflect the new saved data in the table
    roleCanvasMap.value.set(editingRole.value.roleId, [...selectedCanvases.value]);
    toast.success(`Access updated for ${editingRole.value.roleName}`);
    closeEditModal();
  } else {
    toast.error(updateError.value?.message || "Failed to update access mapping");
  }
};

onMounted(async () => {
  await setupData();
});
</script>