import { useRouter } from 'next/router';
import Link from 'next/link';
import useSWR from 'swr';
import { formatDateTimeString, humanizeUnderscoreString } from '../../../common/utils';
import { IAppContext, declarePageTitle, declareValidatingFetchedData } from '../../../components/app_context';
import { NavigationSection } from '../../../components/navbar';
import DataRefreshErrorSnackbar from '../../../components/data_refresh_error_snackbar';
import DataLoadErrorScreen from '../../../components/data_load_error_screen';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import Skeleton from '@material-ui/lab/Skeleton';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableRow from '@material-ui/core/TableRow';
import styles from '../../../common/tables.module.scss';

interface IProps {
  appContext: IAppContext;
}

export default function ReleasePage(props: IProps) {
  const { appContext } = props;
  const router = useRouter();
  const applicationId = router.query.application_id as string;
  const id = router.query.id as string;
  const { data: appData, error: appError, isValidating: appDataIsValidating, mutate: appDataMutate } =
    useSWR(`/v1/applications/${encodeURIComponent(applicationId)}`);
  const { data: releaseData, error: releaseError, isValidating: releaseIsValidating, mutate: releaseMutate } =
    useSWR(`/v1/applications/${encodeURIComponent(applicationId)}/releases/${encodeURIComponent(id)}`);
  const hasAllData = appData && releaseData;
  const firstError = appError || releaseError;
  const isValidating = appDataIsValidating || releaseIsValidating;

  declarePageTitle(appContext, `Release: ${applicationId}/${id}`);
  declareValidatingFetchedData(appContext, isValidating);

  function mutateAll() {
    appDataMutate();
    releaseMutate();
  }

  if (hasAllData) {
    return (
      <>
        <DataRefreshErrorSnackbar error={firstError} refreshing={isValidating} onReload={mutateAll} />
        <Box mx={2} my={2} style={{ flexGrow: 1 }}>
          <TableContainer component={Paper} className={styles.definition_list_table}>
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell component="th" scope="row">ID</TableCell>
                  <TableCell>
                    <Link href={`/releases/${encodeURIComponent(appData.id)}/${encodeURIComponent(releaseData.id)}`}>
                      <a>{releaseData.id}</a>
                    </Link>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Application</TableCell>
                  <TableCell>
                    <Link href={`/applications/${encodeURIComponent(appData.id)}`}>
                      <a>{appData.display_name}</a>
                    </Link>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">State</TableCell>
                  <TableCell>{humanizeUnderscoreString(releaseData.state as string)}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Created at</TableCell>
                  <TableCell>{formatDateTimeString(releaseData.created_at as string)}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Finalized at</TableCell>
                  <TableCell>{releaseData.finalized_at ? formatDateTimeString(releaseData.finalized_at as string) : 'N/A'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Source identity</TableCell>
                  <TableCell>{releaseData.source_identity || 'N/A'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Comments</TableCell>
                  <TableCell>{releaseData.comments || 'N/A'}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>


        </Box>
      </>
    );
  }

  if (firstError) {
    return <DataLoadErrorScreen error={firstError} onReload={mutateAll} />;
  }

  return (
    <Box mx={2} my={2} style={{ display: 'flex', flexDirection: 'column', flexGrow: 1 }}>
      <Paper>
        <Box mx={2} my={2}>
          <Skeleton />
          <Skeleton />
          <Skeleton />
          <Skeleton />
          <Skeleton />
          <Skeleton />
          <Skeleton />
        </Box>
      </Paper>
    </Box>
  );
}

ReleasePage.navigationSection = NavigationSection.Releases;
ReleasePage.pageTitle = 'Release';
ReleasePage.hasBackButton = true;
