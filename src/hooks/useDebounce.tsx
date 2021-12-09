import { useEffect, useState } from 'react'

export const useDebounce = <T,>(value: T, delay: number, initialValue?: T) => {
  const [debouncedValue, setDebouncedValue] = useState<T | undefined>(
    initialValue,
  )

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => clearTimeout(handler)
  }, [value, delay])

  return debouncedValue
}
