import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useLiveStreamStore = defineStore('liveStream', () => {
  const liveData = ref({}); 
  const activeDeviceIds = ref([]);
  let eventSource = null;

  const setRequiredDevices = (deviceIds) => {
    // Convert IDs to strings to ensure strict comparison matches safely
    const uniqueIds = [...new Set(deviceIds)].map(String).sort();
    const currentIds = [...activeDeviceIds.value].map(String).sort();
    
    if (JSON.stringify(uniqueIds) === JSON.stringify(currentIds)) {
      return; 
    }
    
    activeDeviceIds.value = uniqueIds;
    restartSSE();
  };

  const restartSSE = () => {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    if (activeDeviceIds.value.length === 0) {
      liveData.value = {};
      return;
    }

    const deviceIdQuery = activeDeviceIds.value.join(',');
    const baseUrl = import.meta.env.VITE_API_BASE_URL;

    eventSource = new EventSource(`${baseUrl}/device/chartstream?deviceId=${deviceIdQuery}`, {
      withCredentials: true
    });

    eventSource.onmessage = (event) => {
      const payload = JSON.parse(event.data);
      const deviceData = payload.deviceData || {};
      
      const nextMap = { ...liveData.value };
      
      // ⚡ THE FIX: Use Object.entries to capture the backend's exact ID key!
      Object.entries(deviceData).forEach(([key, device]) => {
        // This guarantees the map key matches the actual device ID
        const id = String(device.deviceId || key); 
        
        nextMap[id] = {
          name: device.deviceName || `Device ${id}`,
          value: device.valueData
        };
      });
      
      liveData.value = nextMap;
    };
  };

  const disconnect = () => {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }
    liveData.value = {};
  };

  return { liveData, setRequiredDevices, disconnect };
});