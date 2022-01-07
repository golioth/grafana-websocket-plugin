import { DataSourcePluginOptionsEditorProps } from '@grafana/data'
import { LegacyForms } from '@grafana/ui'
import React, { ChangeEvent, FC } from 'react'
import { DataSourceOptions, SecureJsonData } from '../types'
import CustomHeadersSettings from './fields/CustomHeadersField'
import CustomQueryParamsSettings from './fields/CustomQueryParamsField'

const { SecretFormField, FormField } = LegacyForms

type Props = DataSourcePluginOptionsEditorProps<DataSourceOptions>

export const ConfigEditor: FC<Props> = ({ options, onOptionsChange }) => {
  const { jsonData, secureJsonFields } = options
  const secureJsonData = (options.secureJsonData || {}) as SecureJsonData
  const onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const jsonData = {
      ...options.jsonData,
      host: event.target.value,
    }
    onOptionsChange({ ...options, jsonData })
  }

  // Secure field (only sent to the backend)
  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiKey: event.target.value,
      },
    })
  }

  const onResetAPIKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    })
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
            value={jsonData.host || ''}
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
