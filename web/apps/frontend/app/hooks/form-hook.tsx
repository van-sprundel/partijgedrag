import { useState, type ChangeEvent} from 'react';

export function useFormHook<T>(initialState: T) {
  const [state, setState] = useState<T>(initialState);

  const handleChange = (key: keyof T) => {
    return (e: ChangeEvent<HTMLInputElement>) => {
      setState((prev: T) => ({
        ...prev,
        [key]: e.target.value,
      }));
    };
  };

  return [state, handleChange] as const;
}