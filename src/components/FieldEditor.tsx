import { FieldType, SelectableValue } from '@grafana/data'
import { Icon, InlineField, InlineFieldRow, Select } from '@grafana/ui'
import React from 'react'
import { QueryField, QueryLanguage } from 'types'
import { AliasField } from './fields/AliasField'
import { JsonPathField } from './fields/JsonPathField'

interface Props {
  limit?: number
  onChange: (value: QueryField[]) => void
  value: QueryField[]
}

export const FieldEditor = ({ value = [], onChange, limit }: Props) => {
  const onChangePath = (i: number) => (e: string) => {
    onChange(
      value.map((field, n) => (i === n ? { ...value[i], jsonPath: e } : field)),
    )
  }

  const onLanguageChange = (i: number) => (
    e: SelectableValue<QueryLanguage>,
  ) => {
    onChange(
      value.map((field, n) =>
        i === n ? { ...value[i], language: e.value } : field,
      ),
    )
  }
  const onChangeType = (i: number) => (e: SelectableValue<string>) => {
    onChange(
      value.map((field, n) =>
        i === n
          ? {
              ...value[i],
              type: (e.value === 'auto' ? undefined : e.value) as FieldType,
            }
          : field,
      ),
    )
  }
  const onAliasChange = (i: number) => (e: string) => {
    onChange(
      value.map((field, n) => (i === n ? { ...value[i], name: e } : field)),
    )
  }

  const addField = (
    i: number,
    defaults?: { language: QueryLanguage },
  ) => () => {
    if (!limit || value.length < limit) {
      onChange([
        ...value.slice(0, i + 1),
        { name: '', jsonPath: '', ...defaults },
        ...value.slice(i + 1),
      ])
    }
  }
  const removeField = (i: number) => () => {
    onChange([...value.slice(0, i), ...value.slice(i + 1)])
  }

  return (
    <>
      {value.map((field, index) => (
        <InlineFieldRow key={index}>
          <InlineField
            label='Field'
            tooltip={
              <div>
                A{' '}
                <a
                  href='https://goessner.net/articles/JsonPath/'
                  target='_blank'
                  rel='noreferrer'
                >
                  JSON Path
                </a>{' '}
                query that selects one or more values from a JSON object.
              </div>
            }
            grow
          >
            <JsonPathField
              onChange={onChangePath(index)}
              jsonPath={field.jsonPath}
            />
          </InlineField>
          <InlineField disabled>
            <Select
              width={14}
              value={'jsonpath'}
              onChange={onLanguageChange(index)}
              options={[{ label: 'JSONPath', value: 'jsonpath' }]}
            />
          </InlineField>
          <InlineField
            label='Type'
            tooltip='If Auto is set, the JSON property type is used to detect the field type.'
          >
            <Select
              value={field.type ?? 'auto'}
              width={12}
              onChange={onChangeType(index)}
              options={[
                { label: 'Auto', value: 'auto' },
                { label: 'String', value: 'string' },
                { label: 'Number', value: 'number' },
                { label: 'Time', value: 'time' },
                { label: 'Boolean', value: 'boolean' },
              ]}
            />
          </InlineField>
          <InlineField
            label='Alias'
            tooltip='If left blank, the field uses the name of the queried element.'
          >
            <AliasField onChange={onAliasChange(index)} alias={field.name} />
          </InlineField>

          {(!limit || value.length < limit) && (
            <a
              className='gf-form-label'
              onClick={addField(index, {
                language: field.language ?? 'jsonpath',
              })}
            >
              <Icon name='plus' />
            </a>
          )}

          {value.length > 1 ? (
            <a className='gf-form-label' onClick={removeField(index)}>
              <Icon name='minus' />
            </a>
          ) : null}
        </InlineFieldRow>
      ))}
    </>
  )
}
