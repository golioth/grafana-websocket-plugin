import { DataQuery, DataSourceJsonData, FieldType } from '@grafana/data'

export type QueryLanguage = 'jsonpath'

export interface QueryField {
  name?: string
  jsonPath: string
  type?: FieldType
  language?: QueryLanguage
}

export type Pair<T, K> = [T, K]

export interface Query extends DataQuery {
  path: string
  withStreaming: boolean
  fields: QueryField[]
}

export const defaultQuery: Partial<Query> = {
  path: '',
  fields: [{ jsonPath: '', language: 'jsonpath', name: '' }],
  withStreaming: true,
}

export interface DataSourceOptions extends DataSourceJsonData {
  host?: string
}

export interface SecureJsonData {
  apiKey?: string
}
