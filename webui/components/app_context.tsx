import React, { useLayoutEffect } from 'react';

export interface IAppContext {
  isValidatingFetchedData: boolean,
  setValidatingFetchedData: (val: boolean) => void;
};

export const AppContext = React.createContext({} as IAppContext);

export function declareValidatingFetchedData(appContext: IAppContext, value: boolean) {
  useLayoutEffect(function() {
    appContext.setValidatingFetchedData(value);
  }, [value]);
}
