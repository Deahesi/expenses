import { AppProviders } from './providers'
import { ExpensesPage } from '@/pages/expenses/ui/expenses-page'

export function App() {
  return (
    <AppProviders>
      <ExpensesPage />
    </AppProviders>
  )
}
