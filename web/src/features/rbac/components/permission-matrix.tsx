'use client'

import { useMemo } from 'react'
import {
  ALL_PERMISSION_IDS,
  RBAC_MODULE_GROUPS,
  type ModuleGroup,
} from '../data/catalog'
import { Checkbox } from '@/components/ui/checkbox'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { CheckCheck, RotateCcw } from 'lucide-react'

interface PermissionMatrixProps {
  selectedPermissions: string[]
  onChange: (permissions: string[]) => void
  disabled?: boolean
}

export function PermissionMatrix({
  selectedPermissions,
  onChange,
  disabled = false,
}: PermissionMatrixProps) {
  const selectedSet = useMemo(
    () => new Set(selectedPermissions),
    [selectedPermissions]
  )

  const isAllSelected = useMemo(
    () =>
      ALL_PERMISSION_IDS.length > 0 &&
      ALL_PERMISSION_IDS.every((id) => selectedSet.has(id)),
    [selectedSet]
  )

  const togglePermission = (id: string) => {
    if (disabled) return
    const next = new Set(selectedSet)
    if (next.has(id)) {
      next.delete(id)
    } else {
      next.add(id)
    }
    onChange(Array.from(next))
  }

  const toggleModule = (group: ModuleGroup) => {
    if (disabled) return
    const groupIds = group.permissions.map((p) => p.id)
    const allGroupSelected = groupIds.every((id) => selectedSet.has(id))
    const next = new Set(selectedSet)

    if (allGroupSelected) {
      groupIds.forEach((id) => next.delete(id))
    } else {
      groupIds.forEach((id) => next.add(id))
    }
    onChange(Array.from(next))
  }

  const selectAll = () => {
    if (disabled) return
    onChange([...ALL_PERMISSION_IDS])
  }

  const deselectAll = () => {
    if (disabled) return
    onChange([])
  }

  return (
    <div className='space-y-4'>
      {/* Top Action Bar */}
      <div className='flex flex-wrap items-center justify-between gap-2 rounded-lg border bg-muted/30 px-3 py-2 text-xs'>
        <div className='flex items-center gap-2'>
          <span className='font-medium text-foreground'>
            Permissions Selected:
          </span>
          <Badge variant='secondary' className='font-mono font-semibold'>
            {selectedSet.size} / {ALL_PERMISSION_IDS.length}
          </Badge>
        </div>

        <div className='flex items-center gap-2'>
          {!isAllSelected ? (
            <Button
              type='button'
              variant='outline'
              size='sm'
              className='h-7 text-xs gap-1'
              disabled={disabled}
              onClick={selectAll}
            >
              <CheckCheck className='h-3.5 w-3.5' /> Select All
            </Button>
          ) : (
            <Button
              type='button'
              variant='outline'
              size='sm'
              className='h-7 text-xs gap-1'
              disabled={disabled}
              onClick={deselectAll}
            >
              <RotateCcw className='h-3.5 w-3.5' /> Deselect All
            </Button>
          )}
        </div>
      </div>

      {/* Matrix Module Cards */}
      <div className='space-y-3 max-h-[60vh] overflow-y-auto pr-1'>
        {RBAC_MODULE_GROUPS.map((group) => {
          const groupIds = group.permissions.map((p) => p.id)
          const selectedCount = groupIds.filter((id) => selectedSet.has(id)).length
          const isGroupAll = selectedCount === groupIds.length
          const isGroupSome = selectedCount > 0 && selectedCount < groupIds.length

          return (
            <Card key={group.id} className='border border-border/70 shadow-none'>
              <CardHeader className='py-2.5 px-3.5 bg-muted/20 border-b flex flex-row items-center justify-between space-y-0'>
                <div>
                  <CardTitle className='text-xs font-semibold tracking-tight text-foreground'>
                    {group.label}
                  </CardTitle>
                  <CardDescription className='text-[11px] text-muted-foreground'>
                    {group.description}
                  </CardDescription>
                </div>
                <div className='flex items-center gap-2'>
                  <Button
                    type='button'
                    variant='ghost'
                    size='sm'
                    className='h-6 px-2 text-[11px] font-medium text-muted-foreground hover:text-foreground'
                    disabled={disabled}
                    onClick={() => toggleModule(group)}
                  >
                    {isGroupAll ? 'Deselect Group' : 'Select Group'}
                  </Button>
                  <Badge
                    variant={isGroupAll ? 'default' : isGroupSome ? 'secondary' : 'outline'}
                    className='text-[10px] px-1.5 py-0 h-5 font-mono'
                  >
                    {selectedCount}/{groupIds.length}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className='p-3 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2.5'>
                {group.permissions.map((perm) => {
                  const isChecked = selectedSet.has(perm.id)
                  return (
                    <label
                      key={perm.id}
                      className={`flex items-start gap-2.5 p-2 rounded-md border transition-all cursor-pointer select-none text-start ${
                        isChecked
                          ? 'border-primary/50 bg-primary/5 dark:bg-primary/10'
                          : 'border-border/50 hover:bg-muted/40'
                      } ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
                    >
                      <Checkbox
                        id={perm.id}
                        checked={isChecked}
                        disabled={disabled}
                        onCheckedChange={() => togglePermission(perm.id)}
                        className='mt-0.5'
                      />
                      <div className='space-y-0.5 min-w-0'>
                        <div className='flex items-center gap-1.5'>
                          <span className='text-xs font-medium text-foreground'>
                            {perm.label}
                          </span>
                          <code className='text-[9px] px-1 py-0.2 rounded bg-muted text-muted-foreground font-mono'>
                            {perm.action}
                          </code>
                        </div>
                        <p className='text-[10px] text-muted-foreground line-clamp-2 leading-snug'>
                          {perm.description}
                        </p>
                      </div>
                    </label>
                  )
                })}
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  )
}
