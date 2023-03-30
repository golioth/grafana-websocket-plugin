import { DataSourcePluginOptionsEditorProps } from '@grafana/data'
import { LegacyForms } from '@grafana/ui'
import React, { ChangeEvent, FC } from 'react'
import { DataSourceOptions } from '../types'
import CustomHeadersSettings from './fields/CustomHeadersField'
import CustomQueryParamsSettings from './fields/CustomQueryParamsField'

const { FormField } = LegacyForms

type Props = DataSourcePluginOptionsEditorProps<DataSourceOptions>

export const ConfigEditor: FC<Props> = ({ options, onOptionsChange }) => {
  const { url, jsonData } = options

  const onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const newOptions = {
      ...options,
      url: event.target.value,
    }
    onOptionsChange({ ...newOptions, jsonData })
  }

  return (
    <div className='gf-form-group'>
      <h3 className='page-heading'>WebSocket</h3>
      <div className='gf-form-group'>
        <div className='gf-form'>
          <FormField
            label='Host'
            labelWidth={10}
            inputWidth={30}
            onChange={onHostChange}
            value={url || ''}
            placeholder='wss://api.domain.io/v1/ws/'
          />
        </div>
      </div>

      <CustomHeadersSettings
        dataSourceConfig={options}
        onChange={onOptionsChange}
      />

      <CustomQueryParamsSettings
        dataSourceConfig={options}
        onChange={onOptionsChange}
      />
    </div>
  )
}
