/* eslint-disable @typescript-eslint/no-explicit-any */
import {
  ArrayVector,
  DataFrame,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  Field,
  LoadingState,
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
    const { range, scopedVars, panelId, dashboardId, targets } = options
    const replaceWithVars = replace(scopedVars, range)

    const newOptions = Object.assign({}, options)
    newOptions.targets = newOptions.targets.map(({ path, ...rest }) => ({
      ...rest,
      path: replaceWithVars(path),
    }))

    const result = super.query(newOptions)
    const queries = targets.filter(target => !target.hide).map(target => target)

    return result.pipe(
      map(event => {
        if (event.state === 'Streaming') {
          const eventChannel = event.data[0].meta?.channel

          const newFrames = queries
            .filter(query => {
              const queryChannel = replaceWithVars(
                `ds/${query.datasource?.uid}/${query?.path || '.'}`,
              )
              return queryChannel === eventChannel
            })
            .map(query =>
              this.transformFrames(event.data[0], query, scopedVars, range),
            )

          const key = `${dashboardId}/${panelId}/${eventChannel}`
          const eventBindingKey = key || event.key

          const errors = newFrames
            .filter(f => !!f.meta?.custom)
            .map(frame => frame.meta?.custom)

          const errorMsg = errors.length
            ? `Some queries returned an error:
              ${errors
                .map(error => `Query ${error?.refId} - ${error?.error}`)
                .join('\n')}`
            : undefined

          const newEvent = {
            ...event,
            data: newFrames,
            key: queries.length > 1 ? eventBindingKey : event.key,
            state: LoadingState.Streaming,
            error: errorMsg
              ? {
                  message: errorMsg
                    ? 'Streaming error: click to see details'
                    : undefined,
                  data: {
                    message: errorMsg,
                  },
                }
              : undefined,
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
    const { refId } = query
    if (query?.fields?.length === 0) return { ...eventFrame, refId }

    // casted to any to avoid typescript Vector<any> error
    const eventFields = eventFrame.fields as any[]

    const eventValues = eventFields
      .find(f => f.name === 'data')
      ?.values.map((v: string) => {
        try {
          return JSON.parse(v)
        } catch (e) {
          throw new Error(`Invalid JSON: ${v}. Error: ${e}`)
        }
      })

    const newFields = query?.fields
      ?.filter(field => field.jsonPath)
      .map(field => this.transformFields(field, scopedVars, range, eventValues))

    if (eventFields.find(field => field.name === 'error')) {
      const errorMsg = eventFields.find(field => field.name === 'error')
        ?.values[0]

      return {
        ...eventFrame,
        refId,
        fields: newFields,
        meta: {
          ...eventFrame.meta,
          custom: {
            error: errorMsg,
            refId,
          },
        },
      }
    }

    const newFrame = {
      ...eventFrame,
      refId,
      fields: newFields ?? eventFields,
    }

    return newFrame
  }

  transformFields = (
    field: QueryField,
    scopedVars: ScopedVars,
    range: TimeRange,
    eventValues?: Record<string, unknown>[],
  ): Field => {
    const replaceWithVars = replace(scopedVars, range)
    const path = replaceWithVars(field.jsonPath)
    const values = eventValues?.flatMap(json => JSONPath({ path, json })) ?? []

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
