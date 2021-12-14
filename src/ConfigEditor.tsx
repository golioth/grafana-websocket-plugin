import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { LegacyForms } from '@grafana/ui';
import React, { ChangeEvent, FC } from 'react';
import { DataSourceOptions, SecureJsonData } from './types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<DataSourceOptions> {}

export const ConfigEditor: FC<Props> = ({ options, onOptionsChange }) => {
  const { jsonData, secureJsonFields } = options;
  const secureJsonData = (options.secureJsonData || {}) as SecureJsonData;
  const onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const jsonData = {
      ...options.jsonData,
      host: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiKey: event.target.value,
      },
    });
  };

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
    });
  };

  return (
    <div className="gf-form-group">
      <div className="gf-form">
        <FormField
          label="Host"
          labelWidth={10}
          inputWidth={30}
          onChange={onHostChange}
          value={jsonData.host || ''}
          placeholder="https://localhost:9080"
        />
      </div>

      <div className="gf-form-inline">
        <div className="gf-form">
          <SecretFormField
            isConfigured={(secureJsonFields && secureJsonFields.apiKey) as boolean}
            value={secureJsonData.apiKey || ''}
            label="API Key"
            placeholder="secure field (backend only)"
            labelWidth={10}
            inputWidth={30}
            onReset={onResetAPIKey}
            onChange={onAPIKeyChange}
          />
        </div>
      </div>
    </div>
  );
};
