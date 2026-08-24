import { useId } from 'react'
import { Input } from '@/components/ui/input'

interface ResourceDatalistInputProps {
  value?: string
  onChange: (v: string) => void
  placeholder?: string
  options?: string[] // suggestion dari router aktif
  optionsLoading?: boolean
}

// Input teks bebas + native <datalist> suggestion dari router aktif.
// Tanpa device / gagal fetch: tetap input biasa (fallback aman lintas-router).
export function ResourceDatalistInput({
  value,
  onChange,
  placeholder,
  options = [],
  optionsLoading,
}: ResourceDatalistInputProps) {
  const autoId = useId()
  const listId = `datalist-${autoId}`
  return (
    <>
      <Input
        list={listId}
        value={value ?? ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder={optionsLoading ? 'Memuat dari router…' : (placeholder ?? '')}
        autoComplete='off'
      />
      <datalist id={listId}>
        {options.map((opt) => (
          <option key={opt} value={opt} />
        ))}
      </datalist>
    </>
  )
}
