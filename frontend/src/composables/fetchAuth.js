export const fetchWithAuth = async (url, options = {}) => {
  // ⚡ Check if the body is FormData. If it is, DO NOT set the Content-Type.
  const isFormData = options.body instanceof FormData;

  const headers = {
    // Only apply application/json if it's a standard request
    ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
    ...options.headers
  }

  const response = await fetch(url, { 
    ...options, 
    headers,
    // This tells the browser: "Attach the HttpOnly cookie to this request automatically"
    credentials: 'include' 
  })

  if (response.status === 401) {
    localStorage.removeItem('userId')
    localStorage.removeItem('userName')
    window.location.href = '/login'
    throw new Error('Session expired. Please log in again.')
  }

  return response
}