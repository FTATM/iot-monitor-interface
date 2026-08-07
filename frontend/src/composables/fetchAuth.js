export const fetchWithAuth = async (url, options = {}) => {
  const headers = {
    'Content-Type': 'application/json',
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