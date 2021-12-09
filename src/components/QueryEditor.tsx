import { QueryEditorProps } from '@grafana/data'
import { DataSource } from 'datasource'
import React from 'react'
import { DataSourceOptions, Query } from '../types'
import { FieldEditor } from './FieldEditor'
import { TabbedQueryEditor } from './TabbedQueryEditor'

interface Props extends QueryEditorProps<DataSource, Query, DataSourceOptions> {
  limitFields?: number
  editorContext?: string
}

export const QueryEditor: React.FC<Props> = props => {
  const { query, editorContext, onChange, onRunQuery } = props

  return (
    <TabbedQueryEditor
      {...props}
      editorContext={editorContext || 'default'}
      fieldsTab={
        <FieldEditor
          value={query.fields}
          onChange={value => {
            onChange({ ...query, fields: value })
            onRunQuery()
          }}
          limit={props.limitFields}
        />
      }
    />
  )
}
