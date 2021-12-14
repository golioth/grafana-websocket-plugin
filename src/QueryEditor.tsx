import { defaults } from 'lodash'

import React, { ChangeEvent, FC, SyntheticEvent } from 'react'
import { LegacyForms } from '@grafana/ui'
import { QueryEditorProps } from '@grafana/data'
import { DataSource } from './datasource'
import { DataSourceOptions, Query } from './types'

const { FormField, Switch } = LegacyForms

type Props = QueryEditorProps<DataSource, Query, DataSourceOptions>

export const QueryEditor: FC<Props> = ({
  query: storedQuery,
  onChange,
  onRunQuery,
}) => {
  const defaultQuery: Partial<Query> = {
    withStreaming: false,
    path: '',
  }
  const query = defaults(storedQuery, defaultQuery)
  const { path, withStreaming } = query

  const onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, path: event.target.value })
  }

  const onWithStreamingChange = (event: SyntheticEvent<HTMLInputElement>) => {
    onChange({ ...query, withStreaming: event.currentTarget.checked })
    // executes the query
    onRunQuery()
  }

  return (
    <div className='gf-form'>
      <FormField
        labelWidth={8}
        value={path || ''}
        onChange={onQueryTextChange}
        label='Path'
        tooltip='Websocket URL Path to connect'
        placeholder='/api/v1/ws/realtime'
      />
      <Switch
        checked={withStreaming || false}
        label='Enable streaming (v8+)'
        onChange={onWithStreamingChange}
      />
    </div>
  )
}
