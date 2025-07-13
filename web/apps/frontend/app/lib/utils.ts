import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function capitalize(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export function hasNonNullable<T, K extends keyof T>(key: K) {
  return (obj: T): obj is T & Required<Pick<T, K>> => {
    return obj[key] != null && obj[key] !== undefined;
  };
}