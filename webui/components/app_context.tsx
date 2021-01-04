import React from 'react';
import { useIsomorphicLayoutEffect } from '../common/utils';

export interface IAppContext {
  isValidatingFetchedData: boolean,
  setValidatingFetchedData: (val: boolean) => void;
};

export const AppContext = React.createContext({} as IAppContext);

export function declareValidatingFetchedData(appContext: IAppContext, value: boolean) {
  useIsomorphicLayoutEffect(function() {
    appContext.setValidatingFetchedData(value);
  }, [value]);
}
