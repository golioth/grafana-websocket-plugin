import { Input } from '@grafana/ui'
import { useDebounce } from 'hooks/useDebounce'
import React, { useEffect, useState } from 'react'

interface Props {
  path?: string
  onChange: (v: string) => void
}

export const PathField: React.FC<Props> = ({ path, onChange }) => {
  const [value, setValue] = useState(path)
  const debouncedValue = useDebounce(value, 500, path)

  useEffect(() => {
    if (debouncedValue === path) return

    onChange(debouncedValue || '')
  }, [debouncedValue])

  return (
    <Input
      value={value}
      label='Path'
      placeholder='/api/v1/ws/realtime'
      onChange={e => setValue(e.currentTarget.value)}
    />
  )
}
