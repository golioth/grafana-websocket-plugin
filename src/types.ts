import { DataQuery, DataSourceJsonData } from '@grafana/data'

export interface Query extends DataQuery {
  path?: string
  withStreaming: boolean
}

/**
 * These are options configured for each DataSource instance.
 */
export interface DataSourceOptions extends DataSourceJsonData {
  host?: string
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SecureJsonData {
  apiKey?: string
}
