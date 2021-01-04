import { useContext, useState } from 'react';
import { useRouter, NextRouter } from 'next/router';
import Head from 'next/head';
import { SWRConfig } from 'swr';
import axios from 'axios';
import Layout from '../components/layout';
import { IAppContext, AppContext } from '../components/app_context';
import { IUser } from '../common/user';
import { createMuiTheme, ThemeProvider } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import blue from '@material-ui/core/colors/blue';
import { useIsomorphicLayoutEffect } from '../common/utils';
import '../common/global.css';

const THEME = createMuiTheme({
  palette: {
    primary: { main: blue[500] }
  },
});

const AXIOS = axios.create({
  baseURL: 'http://localhost:3001',
  headers: {
    'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTI5Mzk4NjYsIm9taWQiOiJhZG1pbl9zYSIsIm9tdCI6InNhIiwib3JnaWQiOiJvcmcxIiwib3JpZ19pYXQiOjE2MDkyNTM0NjZ9.mIrFXNytnbAOgjdx_1U2WLqwqE_n6yYq-eFGv3e7Kf0'
  }
});

const SWR_OPTIONS = {
  fetcher: (url: string) => AXIOS.get(url).then(res => res.data),
  revalidateOnFocus: false,
  revalidateOnReconnect: false,
  shouldRetryOnError: false,
}

function useEffect_HideProgressBarOnRouteChange(router: NextRouter, appContext: IAppContext) {
  useIsomorphicLayoutEffect(function() {
    function handleRouteChange() {
      if (appContext.isValidatingFetchedData) {
        appContext.setValidatingFetchedData(false);
      }
    }

    router.events.on('routeChangeStart', handleRouteChange);
    return function() {
      router.events.off('routeChangeStart', handleRouteChange);
    }
  }, [appContext]);
}

function AppWithContext({ Component, pageProps }) {
  const appContext = useContext(AppContext);
  const router = useRouter();
  const pageTitle = Component.pageTitle || 'Sqedule';
  const documentTitle = Component.pageTitle ? `${Component.pageTitle} — Sqedule` : 'Sqedule';
  const user: IUser = { full_name: 'Hongli' };

  useEffect_HideProgressBarOnRouteChange(router, appContext);

  return (
    <>
      <Head>
        <title>{documentTitle}</title>
        <link rel="stylesheet" href="//fonts.googleapis.com/css?family=Roboto:300,400,500,700&amp;display=swap" />
      </Head>

      <CssBaseline />

      <Layout title={pageTitle} loading={appContext.isValidatingFetchedData} user={user}>
        <Component appContext={appContext} {...pageProps} />
      </Layout>
    </>
  );
}

export default function App({ Component, pageProps }) {
  const [isValidatingFetchedData, setValidatingFetchedData] = useState(false);
  const appContextValue: IAppContext = {
    isValidatingFetchedData: isValidatingFetchedData,
    setValidatingFetchedData: setValidatingFetchedData,
  };

  return (
    <SWRConfig value={SWR_OPTIONS}>
      <ThemeProvider theme={THEME}>
        <AppContext.Provider value={appContextValue}>
          <AppWithContext
            Component={Component}
            pageProps={pageProps} />
        </AppContext.Provider>
      </ThemeProvider>
    </SWRConfig>
  )
}
