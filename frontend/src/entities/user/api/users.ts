import { api } from '@/shared/api/client'
import type { User } from '../model/types'

export async function getUsers(signal?: AbortSignal) {
  const { data } = await api.get<User[]>('/users/', { signal })

  return data
}
