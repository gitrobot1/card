export async function parseResponse<T>(response: Response): Promise<T> {
  const text = await response.text()
  if (!text) {
    if (!response.ok) {
      throw new Error(`请求失败 (${response.status})`)
    }
    return {} as T
  }

  try {
    return JSON.parse(text) as T
  } catch {
    if (response.status === 404) {
      throw new Error('接口不存在，请重启 backend 后再试')
    }
    throw new Error(text.slice(0, 120) || `响应格式错误 (${response.status})`)
  }
}

export function readApiError(data: unknown, fallback: string) {
  if (data && typeof data === 'object' && 'error' in data) {
    const message = (data as { error?: unknown }).error
    if (typeof message === 'string' && message) {
      return message
    }
  }
  return fallback
}
