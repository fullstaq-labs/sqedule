import React from 'react';
import { useIsomorphicLayoutEffect } from '../common/utils';

export interface IAppContext {
  pageTitle: string;
  setPageTitle: (val: string) => void;

  isValidatingFetchedData: boolean,
  setValidatingFetchedData: (val: boolean) => void;
};

export const AppContext = React.createContext({} as IAppContext);

export function declarePageTitle(appContext: IAppContext, value: string) {
  useIsomorphicLayoutEffect(function() {
    appContext.setPageTitle(value);
  }, [value]);
}

export function declareValidatingFetchedData(appContext: IAppContext, value: boolean) {
  useIsomorphicLayoutEffect(function() {
    appContext.setValidatingFetchedData(value);
  }, [value]);
}
