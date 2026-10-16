import { categoryMeta } from '../model/category-meta'
import type { ExpenseCategory } from '../model/types'
import { colorByTextHash } from '@/shared/lib/color-by-text-hash'

interface CategoryBadgeProps {
  category: ExpenseCategory
}

export function CategoryBadge({ category }: CategoryBadgeProps) {
  return (
    <span style={{ color: colorByTextHash(category) }}>
      {categoryMeta[category].label}
    </span>
  )
}
