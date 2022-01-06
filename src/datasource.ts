/* eslint-disable @typescript-eslint/no-explicit-any */
import {
  ArrayVector,
  DataFrame,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  Field,
  ScopedVars,
  TimeRange,
} from '@grafana/data'
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime'
import { detectFieldType } from 'detectFieldType'
import { JSONPath } from 'jsonpath-plus'
import { parseValues } from 'parseValues'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { DataSourceOptions, Query, QueryField } from './types'

export class DataSource extends DataSourceWithBackend<
  Query,
  DataSourceOptions
> {
  constructor(instanceSettings: DataSourceInstanceSettings<DataSourceOptions>) {
    super(instanceSettings)
  }

  query = (options: DataQueryRequest<Query>): Observable<DataQueryResponse> => {
    const result = super.query(options)
    const queries = options.targets
      .filter(target => !target.hide)
      .map(target => target)

    const { range, scopedVars, panelId, dashboardId } = options

    return result.pipe(
      map(event => {
        if (event.state === 'Streaming') {
          const eventChannel = event.data[0].meta?.channel
          const newFrames = queries
            .filter(query => {
              const queryChannel = `ds/${query.datasource?.uid}/${query.path}`
              return queryChannel === eventChannel
            })
            .map(query => {
              return this.transformFrames(
                event.data[0],
                query,
                scopedVars,
                range,
              )
            })

          const key = `${dashboardId}/${panelId}/${eventChannel}`
          const eventBindingKey = key || event.key

          const newEvent = {
            ...event,
            data: newFrames,
            key: queries.length > 1 ? eventBindingKey : event.key,
          }
          return newEvent
        }

        return event
      }),
    )
  }

  transformFrames = (
    eventFrame: DataFrame,
    query: Query,
    scopedVars: ScopedVars,
    range: TimeRange,
  ): DataFrame => {
    // casted to any to avoid typescript Vector[] error
    const eventFields = eventFrame.fields as any[]
    const eventValues = eventFields
      .find(f => f.type !== 'time')
      ?.values.map((v: string) => {
        try {
          JSON.parse(v)
        } catch (e) {
          throw new Error(`Invalid JSON: ${v}`)
        }
      })

    if (!eventValues || query?.fields.length === 0)
      return { ...eventFrame, refId: query.refId }

    if (eventValues.error)
      throw new Error('Something went wrong: ' + eventValues.error)

    const newFields = query?.fields
      .filter(field => field.jsonPath)
      .map(field => this.transformFields(field, eventValues, scopedVars, range))

    const newFrame = {
      ...eventFrame,
      refId: query.refId,
      fields: newFields ?? eventFrame.fields,
    }

    return newFrame
  }

  transformFields = (
    field: QueryField,
    eventValues: Record<string, unknown>[],
    scopedVars: ScopedVars,
    range: TimeRange,
  ): Field => {
    const replaceWithVars = replace(scopedVars, range)
    const path = replaceWithVars(field.jsonPath)
    const values = eventValues.flatMap(json => JSONPath({ path, json }))

    // Get the path for automatic setting of the field name.
    const paths = JSONPath.toPathArray(path)

    const propertyType = field.type ? field.type : detectFieldType(values)
    const typedValues = parseValues(values, propertyType)

    return {
      name: replaceWithVars(field.name ?? '') || paths[paths.length - 1],
      type: propertyType,
      values: new ArrayVector(typedValues),
      config: {},
    }
  }
}

// const getQueryByRefId = (queries: Query[], refId?: string) => {
//   if (refId === undefined) return queries[0]
//   return queries.find(query => query.refId === refId)
// }

const replace = (scopedVars?: ScopedVars, range?: TimeRange) => (
  str: string,
): string => {
  return replaceMacros(getTemplateSrv().replace(str, scopedVars), range)
}

// replaceMacros substitutes all available macros with their current value.
export const replaceMacros = (str: string, range?: TimeRange) => {
  return range
    ? str
        .replace(/\$__unixEpochFrom\(\)/g, range.from.unix().toString())
        .replace(/\$__unixEpochTo\(\)/g, range.to.unix().toString())
        .replace(/\$__isoFrom\(\)/g, range.from.toISOString())
        .replace(/\$__isoTo\(\)/g, range.to.toISOString())
    : str
}
