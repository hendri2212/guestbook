const defaultApiBaseUrl = `${window.location.protocol}//${window.location.hostname}:8087`

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || defaultApiBaseUrl
