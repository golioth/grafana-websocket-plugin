import { defaults } from 'lodash';

import React, { ChangeEvent, PureComponent, SyntheticEvent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';

const { FormField, Switch } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, path: event.target.value });
  };

  onWithStreamingChange = (event: SyntheticEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, withStreaming: event.currentTarget.checked });
    // executes the query
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { path, withStreaming } = query;

    return (
      <div className="gf-form">
        <FormField
          labelWidth={8}
          value={path || ''}
          onChange={this.onPathChange}
          label="Path"
          tooltip="Websocket URL Path to connect"
          placeholder="/api/v1/ws/realtime"
        />
        <Switch checked={withStreaming || false} label="Enable streaming (v8+)" onChange={this.onWithStreamingChange} />
      </div>
    );
  }
}
