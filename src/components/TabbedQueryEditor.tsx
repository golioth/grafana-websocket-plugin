import { TimeRange } from '@grafana/data'
import { InlineField, InlineFieldRow, RadioButtonGroup } from '@grafana/ui'
import { DataSource } from 'datasource'
import defaults from 'lodash/defaults'
import React, { useState } from 'react'
import { defaultQuery, Query } from '../types'
import { PathField } from './fields/PathField'

interface Props {
  onChange: (query: Query) => void
  onRunQuery: () => void
  editorContext: string
  query: Query
  limitFields?: number
  datasource: DataSource
  range?: TimeRange
  fieldsTab: React.ReactNode
}

export const TabbedQueryEditor = ({
  query,
  onChange,
  onRunQuery,
  fieldsTab,
}: Props) => {
  const [tabIndex, setTabIndex] = useState(0)

  const q = defaults(query, defaultQuery)

  const onChangePath = (value: string) => {
    onChange({ ...q, path: value })
    onRunQuery()
  }

  const tabs = [
    {
      title: 'Fields',
      content: fieldsTab,
    },
    {
      title: 'Path',
      content: (
        <InlineField label='Path' tooltip='Websocket URL Path to connect.'>
          <PathField path={q.path} onChange={onChangePath} />
        </InlineField>
      ),
    },
  ]

  return (
    <>
      <InlineFieldRow>
        <InlineField>
          <RadioButtonGroup
            onChange={e => setTabIndex(e ?? 0)}
            value={tabIndex}
            options={tabs.map((tab, idx) => ({ label: tab.title, value: idx }))}
          />
        </InlineField>
      </InlineFieldRow>
      {tabs[tabIndex].content}
    </>
  )
}
