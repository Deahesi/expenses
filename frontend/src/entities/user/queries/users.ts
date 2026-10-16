import { queryOptions } from '@tanstack/react-query'
import { getUsers } from '../api/users'

export const userKeys = { all: ['users'] as const }

export function usersQuery() {
  return queryOptions({
    queryKey: userKeys.all,
    queryFn: ({ signal }) => getUsers(signal),
  })
}
