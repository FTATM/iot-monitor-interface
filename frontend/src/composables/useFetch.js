// src/composables/useFetch.js
import { ref } from 'vue'
import { fetchWithAuth } from '../composables/fetchAuth'

export function useFetch() {
  const data = ref(null)
  const isLoading = ref(false)
  const error = ref(null)
  const res = ref(null)

  const execute = async (url) => {
    isLoading.value = true
    error.value = null
    try {
      // Use the wrapper instead of raw fetch
      const response = await fetchWithAuth(url, { method: 'GET' }) 
      res.value = response;
      
      if (!response.ok) throw new Error('Network response was not ok')
      data.value = await response.json()
    } catch (err) {
      error.value = err.message
    } finally {
      isLoading.value = false
    }
  }

  return { data, isLoading, error, res, execute }
}