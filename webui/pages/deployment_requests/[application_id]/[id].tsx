import { useRouter } from 'next/router';
import Link from 'next/link';
import useSWR from 'swr';
import { formatDateTimeString, humanizeUnderscoreString } from '../../../common/utils';
import { IAppContext, declarePageTitle, declareValidatingFetchedData } from '../../../components/app_context';
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

export default function DeploymentRequestPage(props: IProps) {
  const { appContext } = props;
  const router = useRouter();
  const applicationId = router.query.application_id as string;
  const id = router.query.id as string;
  const { data, error, isValidating, mutate } = useSWR(`/v1/applications/${encodeURIComponent(applicationId)}/deployment-requests/${encodeURIComponent(id)}`);

  declarePageTitle(appContext, `Deployment request: ${applicationId}/${id}`);
  declareValidatingFetchedData(appContext, isValidating);

  if (data) {
    return (
      <>
        <DataRefreshErrorSnackbar error={error} refreshing={isValidating} onReload={mutate} />
        <Box mx={2} my={2} style={{ flexGrow: 1 }}>
          <TableContainer component={Paper} className={styles.definition_list_table}>
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell component="th" scope="row">ID</TableCell>
                  <TableCell>
                    <Link href={`/deployment_requests/${encodeURIComponent(data.application.id)}/${encodeURIComponent(data.id)}`}>
                      <a>{data.id}</a>
                    </Link>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Application</TableCell>
                  <TableCell>
                    <Link href={`/applications/${encodeURIComponent(data.application.id)}`}>
                      <a>{data.application.display_name}</a>
                    </Link>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">State</TableCell>
                  <TableCell>{humanizeUnderscoreString(data.state as string)}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Created at</TableCell>
                  <TableCell>{formatDateTimeString(data.created_at as string)}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Finalized at</TableCell>
                  <TableCell>{data.finalized_at ? formatDateTimeString(data.finalized_at as string) : 'N/A'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Source identity</TableCell>
                  <TableCell>{data.source_identity || 'N/A'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell component="th" scope="row">Comments</TableCell>
                  <TableCell>{data.comments || 'N/A'}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>


        </Box>
      </>
    );
  }

  if (error) {
    return <DataLoadErrorScreen error={error} onReload={mutate} />;
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

DeploymentRequestPage.pageTitle = 'Deployment request';
DeploymentRequestPage.hasBackButton = true;
