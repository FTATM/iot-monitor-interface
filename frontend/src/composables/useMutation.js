// src/composables/useMutation.js
import { ref } from 'vue'
import { fetchWithAuth } from '@/composables/fetchAuth'

export function useMutation() {
  const data = ref(null)
  const isLoading = ref(false)
  const error = ref(null)
  const res = ref(null)
  const baseUrl = import.meta.env.VITE_API_BASE_URL

  const execute = async (url, payload, method = 'POST') => {
    isLoading.value = true
    error.value = null

    try {
      // Use the wrapper and pass the body
      const response = await fetchWithAuth(`${baseUrl}${url}`, {
        method: method,
        body: JSON.stringify(payload)
      })
      res.value = response;

      if (!response.ok) {
        const errorData = await response.json().catch(() => null)
        error.value = errorData //|| { message: 'Failed to submit data' }
        return false 
      }

      // Handle successful response
      const responseText = await response.text()
      data.value = responseText ? JSON.parse(responseText) : null
      return true

    } catch (err) {
      // This catch block now only handles actual network failures
      // (like the server being completely down or CORS issues)
      error.value = { message: err.message }
      return false
    } finally {
      isLoading.value = false
    }
  }

  return { data, isLoading, error, res, execute }
}