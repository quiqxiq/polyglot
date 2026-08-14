import { useMemo, useState } from 'react'
import { PlusIcon, Cross2Icon } from '@radix-ui/react-icons'
import { useNavigate } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useKnowledgeListQuery } from './api/use-knowledge'
import { KnowledgeDialogs } from './components/knowledge-dialogs'
import { KnowledgeProvider } from './components/knowledge-provider'
import { KnowledgeTable } from './components/knowledge-table'

function KnowledgeContent() {
  const { data: items = [], isLoading } = useKnowledgeListQuery()
  const navigate = useNavigate()

  const [searchTerm, setSearchTerm] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('all')

  const categories = useMemo(() => {
    const set = new Set<string>()
    for (const item of items) {
      if (item.category) set.add(item.category)
    }
    return [...set].sort()
  }, [items])

  const filteredItems = useMemo(() => {
    const q = searchTerm.trim().toLowerCase()
    return items.filter((item) => {
      if (categoryFilter !== 'all' && item.category !== categoryFilter)
        return false
      if (!q) return true
      return (
        item.title.toLowerCase().includes(q) ||
        item.content.toLowerCase().includes(q) ||
        item.category.toLowerCase().includes(q) ||
        (item.tags ?? []).some((tag) => tag.toLowerCase().includes(q))
      )
    })
  }, [items, searchTerm, categoryFilter])

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* ===== Title & Top Toolbar ===== */}
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>
              Knowledge Base
            </h1>
            <p className='text-sm text-muted-foreground'>
              Manage knowledge documents for the admin dashboard and the
              WhatsApp bot vector store.
            </p>
          </div>
          <Button
            onClick={() => navigate({ to: '/knowledge/new' })}
            className='h-9 gap-1.5 text-xs sm:text-sm'
          >
            <PlusIcon className='h-4 w-4' />
            New Document
          </Button>
        </div>

        {/* ===== Filter & Search Controls ===== */}
        <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
          <div className='flex flex-1 flex-wrap items-center gap-2'>
            <div className='relative w-full sm:w-64'>
              <Input
                placeholder='Filter documents by title, content, or tag...'
                className='h-9 pr-8 text-xs sm:text-sm'
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
              {searchTerm && (
                <button
                  type='button'
                  onClick={() => setSearchTerm('')}
                  className='absolute top-1/2 right-2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
                  title='Clear search'
                >
                  <Cross2Icon className='h-3.5 w-3.5' />
                </button>
              )}
            </div>

            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
              <SelectTrigger className='h-9 w-44 text-xs sm:text-sm'>
                <SelectValue placeholder='All Categories' />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='all'>All Categories</SelectItem>
                {categories.map((category) => (
                  <SelectItem
                    key={category}
                    value={category}
                    className='capitalize'
                  >
                    {category}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <Separator className='shadow-xs' />

        {/* ===== Main Table ===== */}
        <KnowledgeTable data={filteredItems} isLoading={isLoading} />
      </Main>

      <KnowledgeDialogs />
    </>
  )
}

export function Knowledge() {
  return (
    <KnowledgeProvider>
      <KnowledgeContent />
    </KnowledgeProvider>
  )
}
