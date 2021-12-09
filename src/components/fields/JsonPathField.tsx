import { BracesPlugin, QueryField, SlatePrism } from '@grafana/ui'
import { useDebounce } from 'hooks/useDebounce'
import React, { useEffect, useState } from 'react'

interface Props {
  jsonPath: string
  onChange: (v: string) => void
}

/**
 * JsonPathQueryField is an editor for JSON Path.
 */
export const JsonPathField: React.FC<Props> = ({ jsonPath, onChange }) => {
  const [value, setValue] = useState(jsonPath)
  const debouncedValue = useDebounce(value, 500, jsonPath)

  useEffect(() => {
    if (debouncedValue === jsonPath) return

    onChange(debouncedValue || '')
  }, [debouncedValue])

  /**
   * The QueryField supports Slate plugins, so let's add a few useful ones.
   */
  const plugins = [
    BracesPlugin(),
    SlatePrism({
      onlyIn: node => node.type === 'code_block',
      getSyntax: () => 'js',
    }),
  ]

  // This is important if you don't want punctuation to interfere with your suggestions.
  const cleanText = (s: string) =>
    s.replace(/[{}[\]="(),!~+\-*/^%\|\$@\.]/g, '').trim()

  return (
    <QueryField
      additionalPlugins={plugins}
      query={value}
      cleanText={cleanText}
      onChange={v => setValue(v)}
      portalOrigin='wsapi'
      placeholder='$.items[*].name'
    />
  )
}
