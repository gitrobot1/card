export interface AppConfig {
  appName: string
  apiBaseUrl: string
  wsBaseUrl: string
}

const defaultConfig: AppConfig = {
  appName: 'Card Hub',
  apiBaseUrl: '',
  wsBaseUrl: '',
}

let cachedConfig: AppConfig | null = null

function resolveConfig(raw: AppConfig): AppConfig {
  if (typeof window === 'undefined') {
    return raw
  }

  const origin = window.location.origin
  const wsOrigin = origin.replace(/^http/, 'ws')

  return {
    ...raw,
    apiBaseUrl: raw.apiBaseUrl || origin,
    wsBaseUrl: raw.wsBaseUrl || `${wsOrigin}/ws`,
  }
}

export async function loadAppConfig(): Promise<AppConfig> {
  if (cachedConfig) {
    return cachedConfig
  }

  try {
    const response = await fetch('/config.json', { cache: 'no-store' })
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    const raw = { ...defaultConfig, ...(await response.json()) }
    cachedConfig = resolveConfig(raw)
  } catch {
    cachedConfig = resolveConfig(defaultConfig)
  }

  return cachedConfig!
}

export function getAppConfig(): AppConfig {
  return cachedConfig ?? resolveConfig(defaultConfig)
}
