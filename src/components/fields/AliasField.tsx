import { Input } from '@grafana/ui'
import { useDebounce } from 'hooks/useDebounce'
import React, { useEffect, useState } from 'react'

interface Props {
  alias?: string
  onChange: (v: string) => void
}

export const AliasField: React.FC<Props> = ({ alias, onChange }) => {
  const [value, setValue] = useState(alias)
  const debouncedValue = useDebounce(value, 500, alias)

  useEffect(() => {
    if (debouncedValue === alias) return

    onChange(debouncedValue || '')
  }, [debouncedValue])

  return (
    <Input
      width={12}
      value={value}
      onChange={e => setValue(e.currentTarget.value)}
    />
  )
}
