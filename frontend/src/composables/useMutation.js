// src/composables/useMutation.js
import { ref } from 'vue'
import { fetchWithAuth } from '../composables/fetchAuth'

export function useMutation() {
  const data = ref(null)
  const isLoading = ref(false)
  const error = ref(null)
  const res = ref(null)

  const execute = async (url,payload, method = 'POST') => {
    isLoading.value = true
    error.value = null
    
    try {
      // Use the wrapper and pass the body
      const response = await fetchWithAuth(url, {
        method: method,
        body: JSON.stringify(payload)
      })
      res.value = response;

      // if (!response.ok) {
      //   const errorText = await response.text()
      //   throw new Error(errorText || 'Failed to submit data')
      // }

      const responseText = await response.text()
      data.value = responseText ? JSON.parse(responseText) : null
      return true 
      
    } catch (err) {
      error.value = err.message
      return false 
    } finally {
      isLoading.value = false
    }
  }

  return { data, isLoading, error, res, execute }
}