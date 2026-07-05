const DEFAULT_DEV_API_BASE_URL = 'http://localhost:8087'
const DEFAULT_PRODUCTION_API_BASE_URL = ''

export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ||
  (import.meta.env.DEV ? DEFAULT_DEV_API_BASE_URL : DEFAULT_PRODUCTION_API_BASE_URL)
