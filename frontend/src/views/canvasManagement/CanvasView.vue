<template>
    <NoAccess v-if="!hasPermission('Canvas', 'Display')" />

    <div v-else class="p-4 sm:p-6 w-full mx-auto">

        <div
            class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
            <div class="flex items-center gap-4">
                <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
                    <Icon icon="lucide:layout-dashboard" class="w-7 h-7" />
                </div>
                <div>
                    <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">Canvas Management</h2>
                    <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">Create, rename, and remove dashboard
                        canvases.</p>
                </div>
            </div>
        </div>

        <!-- Canvas Table -->
        <TableData :data="canvasTable" :columns="tableColumns" :initial-sorting="[{ id: 'canvasId', desc: false }]"
            :is-loading="isLoadingCanvases">
            <!-- Toolbar Action Slot -->
            <template #toolbar-actions>
                <button @click="openCreateModal" class="btn btn-primary text-white">
                    <Icon icon="lucide:plus" class="w-5 h-5 mr-1" /> Create Canvas
                </button>
            </template>

            <!-- Custom Cell Slot: 'canvasId' -->
            <template #cell-canvasId="{ value }">
                <span class="font-medium text-base-content/50">{{ value }}</span>
            </template>

            <!-- Custom Cell Slot: 'actions' -->
            <template #cell-actions="{ row }">
                <div class="flex justify-end gap-2">
                    <button @click="openEditModal(row)" class="btn btn-sm btn-ghost text-primary">
                        <Icon icon="lucide:pencil" class="w-4 h-4" />
                    </button>
                    <button @click="openDeleteModal(row)" class="btn btn-sm btn-ghost text-error">
                        <Icon icon="lucide:trash-2" class="w-4 h-4" />
                    </button>
                </div>
            </template>

        </TableData>

        <!-- Create / Edit Modal -->
        <dialog ref="formModal" class="modal z-[200]">
            <div class="modal-box">
                <h3 class="font-bold text-lg text-base-content mb-4 flex items-center gap-2">
                    <Icon :icon="form.canvasId ? 'lucide:pencil' : 'lucide:plus'" class="w-5 h-5 text-primary" />
                    {{ form.canvasId ? 'Rename Canvas' : 'Create New Canvas' }}
                </h3>

                <div class="form-control w-full">
                    <label class="label pb-1">
                        <span class="label-text font-bold">Canvas Name</span>
                    </label>
                    <input type="text" v-model="form.canvasName" class="input input-bordered w-full"
                        placeholder="e.g., Main Dashboard" @keyup.enter="submitForm" />
                </div>

                <div class="modal-action mt-6">
                    <button type="button" @click="closeFormModal" class="btn btn-ghost"
                        :disabled="isMutating">Cancel</button>
                    <button type="button" @click="submitForm" class="btn btn-primary text-white"
                        :disabled="isMutating || !form.canvasName.trim()">
                        <span v-if="isMutating" class="loading loading-spinner loading-sm"></span>
                        {{ form.canvasId ? 'Save Changes' : 'Create Canvas' }}
                    </button>
                </div>
            </div>
            <form method="dialog" class="modal-backdrop"><button @click="closeFormModal">close</button></form>
        </dialog>

        <!-- Delete Confirmation Modal -->
        <dialog ref="deleteModal" class="modal z-[200]">
            <div class="modal-box">
                <h3 class="font-bold text-lg text-error flex items-center gap-2">
                    <Icon icon="lucide:alert-triangle" class="w-6 h-6" />
                    Delete Canvas
                </h3>
                <p class="py-4 text-base-content/70">
                    Are you sure you want to delete <span class="font-bold text-base-content">"{{
                        canvasToDelete?.canvasName
                    }}"</span>?
                    This action cannot be undone and will remove all widgets configured on this canvas.
                </p>
                <div class="modal-action">
                    <button type="button" @click="closeDeleteModal" class="btn btn-ghost" :disabled="isMutating">No,
                        Cancel</button>
                    <button type="button" @click="confirmDelete" class="btn btn-error text-white"
                        :disabled="isMutating">
                        <span v-if="isMutating" class="loading loading-spinner loading-sm"></span>
                        Yes, Delete
                    </button>
                </div>
            </div>
            <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
        </dialog>

    </div>
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

// --- STORE & COMPOSABLES ---
const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { data: canvasData, isLoading: isLoadingCanvases, error: canvasError, execute: fetchCanvasesApi } = useFetch();
const { res: mutateRes, isLoading: isMutating, error: mutateError, execute: mutateApi } = useMutation();

// --- STATE ---
const canvasTable = ref([]);
const formModal = ref(null);
const deleteModal = ref(null);

const form = ref({
    canvasId: null,
    canvasName: ''
});
const canvasToDelete = ref(null);

const tableColumns = [
    {
        header: 'ID',
        accessorKey: 'canvasId', // The key in your data object
        meta: { headerClass: 'w-16', cellClass: 'font-bold' } // Custom classes
    },
    {
        header: 'Canvas Name',
        accessorKey: 'canvasName'
    },
    {
        header: 'Actions',
        id: 'actions', // Use 'id' when there isn't a direct data key
        enableSorting: false,
        meta: { headerClass: 'text-right', cellClass: 'text-right' }
    }
];

// --- DATA LOADING ---
const loadCanvases = async () => {
    await fetchCanvasesApi('/canvas/getall');
    if (!canvasError.value && canvasData.value) {
        canvasTable.value = canvasData.value.data || [];
    } else {
        toast.error("Failed to load canvases");
    }
};

// --- FORM MODAL ACTIONS (CREATE & UPDATE) ---
const openCreateModal = () => {
    form.value = { canvasId: null, canvasName: '' };
    formModal.value.showModal();
};

const openEditModal = (canvas) => {
    form.value = { canvasId: canvas.canvasId, canvasName: canvas.canvasName };
    formModal.value.showModal();
};

const closeFormModal = () => {
    formModal.value.close();
    form.value = { canvasId: null, canvasName: '' };
};

const submitForm = async () => {
    if (!form.value.canvasName.trim()) return;

    const isUpdate = !!form.value.canvasId;
    const endpoint = isUpdate ? '/canvas/update' : '/canvas/create';

    const httpMethod = isUpdate ? 'PUT' : 'POST';

    const payload = {
        canvasName: form.value.canvasName
    };

    if (isUpdate) {
        payload.canvasId = form.value.canvasId;
    }

    // 2. Pass the dynamic method into your API call
    await mutateApi(endpoint, payload, httpMethod);

    if (!mutateError.value && mutateRes.value?.ok) {
        toast.success(isUpdate ? "Canvas renamed successfully!" : "Canvas created successfully!");
        closeFormModal();
        await loadCanvases(); // Refresh the table
    } else {
        toast.error(mutateError.value?.message || "Failed to save canvas");
    }
};

// --- DELETE MODAL ACTIONS ---
const openDeleteModal = (canvas) => {
    canvasToDelete.value = canvas;
    deleteModal.value.showModal();
};

const closeDeleteModal = () => {
    deleteModal.value.close();
    canvasToDelete.value = null;
};

const confirmDelete = async () => {
    if (!canvasToDelete.value) return;

    const payload = {
        canvasId: canvasToDelete.value.canvasId
    };

    await mutateApi(`/canvas/delete/${canvasToDelete.value.canvasId}`, null, 'DELETE');

    if (!mutateError.value && mutateRes.value?.ok) {
        toast.success("Canvas deleted successfully!");
        closeDeleteModal();
        await loadCanvases(); // Refresh the table
    } else {
        toast.error(mutateError.value?.message || "Failed to delete canvas");
    }
};

onMounted(async () => {
    await loadCanvases();
});
</script>