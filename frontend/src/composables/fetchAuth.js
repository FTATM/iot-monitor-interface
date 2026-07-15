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
    // If the backend says the cookie is expired or invalid, send them to login
    window.location.href = '/login'
    throw new Error('Session expired. Please log in again.')
  }

  return response
}