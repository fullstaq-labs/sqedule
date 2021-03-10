import { useEffect, useLayoutEffect } from 'react';

/**
 * Format a Date object into a human-readable string.
 * This string strives to be readable and unambiguous across different cultures,
 * which is we use a ISO8601-like format.
 */
export function formatDateTime(date: Date): string {
  // https://stackoverflow.com/a/58633686
  return date.toLocaleTimeString( 'sv-SE', {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
  });
}

/**
 * Parses a date time string and formats into a human-readable string.
 * This string strives to be readable and unambiguous across different cultures,
 * which is we use a ISO8601-like format.
 */
export function formatDateTimeString(dateStr: string): string {
  return formatDateTime(new Date(dateStr));
}

/**
 * Given a string, capitalizes its first letter.
 */
export function capitalizeFirstLetter(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

/**
 * Given an identifier in underscore format, such as "in_progress",
 * turns it into a human-friendly format such as "In progress".
 */
export function humanizeUnderscoreString(str: string): string {
  return capitalizeFirstLetter(str.replace('_', ' '));
}

/**
 * Given an error object, returns a string that describes the error.
 *
 * Normally this is taken from `error.message`. But if the error is due
 * to an errored Axios response, then we'll attempt to extract the error
 * message as returned by the server API call.
 */
export function formatErrorMessage(error: any): string {
  if (error.isAxiosError
    && typeof error.response === 'object'
    && error.response !== null
    && typeof error.response.data === 'object'
    && error.response.data !== null
    && typeof error.response.data.error === 'string') {

      return error.response.data.error;
  }

  return error.message;
}

/**
 * React currently throws a warning when using useLayoutEffect during
 * server-side rendering. To get around it, we can conditionally useEffect
 * on the server (no-op) and useLayoutEffect in the browser.
 * https://gist.github.com/gaearon/e7d97cdf38a2907924ea12e4ebdf3c85#gistcomment-2911761
 *
 * We don't care about this warning because we expect Sqedule to be only
 * usable after downloading Javascript anyway.
 */
export const useIsomorphicLayoutEffect =
  (typeof window !== 'undefined')
  ? useLayoutEffect
  : useEffect;

export function paginateArray<T>(ary: Array<T>, page: number, perPage: number): Array<T> {
  const startIndex = (page - 1) * perPage;
  return ary.slice(startIndex, startIndex + perPage + 1);
}
