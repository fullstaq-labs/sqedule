import useSWR from 'swr';
import Link from 'next/link';
import { formatDateTimeString, capitalizeFirstLetter } from '../common/utils';
import { IAppContext, declareValidatingFetchedData } from '../components/app_context';
import DataRefreshErrorSnackbar from '../components/data_refresh_error_snackbar';
import DataLoadErrorScreen from '../components/data_load_error_screen';
import Container from '@material-ui/core/Container';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import Skeleton from '@material-ui/lab/Skeleton';
import { DataGrid, ColDef, ValueFormatterParams } from '@material-ui/data-grid';

interface IProps {
  appContext: IAppContext;
}

const COLUMNS: ColDef[] = [
  {
    field: 'id',
    type: 'number',
    headerName: 'ID',
    width: 120,
    renderCell: ({ row }) => (
      <Box style={{ flexGrow: 1 }}> {/* Make the content properly align right */}
        <Link href={`/applications/${encodeURIComponent(row.application.id)}/deployment_requests/${encodeURIComponent(row.id)}`}>
          <a>{row.id}</a>
        </Link>
      </Box>
    ),
  },
  {
    field: 'application_display_name',
    headerName: 'Application',
    width: 250,
    valueGetter: ({ row }) => row.application.display_name,
    renderCell: ({ row }) => (
      <Link href={`/applications/${encodeURIComponent(row.application.id)}`}>
        <a>{row.application.display_name}</a>
      </Link>
    ),
  },
  {
    field: 'state',
    headerName: 'State',
    valueFormatter: ({ value }) => capitalizeFirstLetter(value as string),
  },
  {
    field: 'created_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Created at',
    valueFormatter: ({ value }) => formatDateTimeString(value as string),
  },
  {
    field: 'finalized_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Finalized at',
    valueFormatter: ({ value }) => formatDateTimeString(value as string),
  },
];

export default function DeploymentRequestsPage(props: IProps) {
  const { appContext } = props;
  const { data, error, isValidating, mutate } = useSWR('/v1/deployment-requests');

  declareValidatingFetchedData(appContext, isValidating);

  if (data) {
    if (data.items.length == 0) {
      return (
        <Container maxWidth="md">
          <Box px={2} py={2} textAlign="center">
            <Typography variant="h5" color="textSecondary">
              There are no deployment requests.
            </Typography>
          </Box>
        </Container>
      );
    }

    return (
      <>
        <DataRefreshErrorSnackbar error={error} refreshing={isValidating} onReload={mutate} />
        <Box mx={2} my={2} style={{ display: 'flex', flexGrow: 1 }}>
          <Paper style={{ display: 'flex', flexGrow: 1 }}>
            <DataGrid rows={data.items} columns={COLUMNS} />
          </Paper>
        </Box>
      </>
    );
  }

  if (error) {
    return <DataLoadErrorScreen error={error} onReload={mutate} />;
  }

  return (
    <Box mx={2} my={2} style={{ display: 'flex', flexGrow: 1 }}>
      <Paper style={{ flexGrow: 1 }}>
        <Box mx={2} my={2}>
          <Container maxWidth="md">
            <Skeleton />
            <Skeleton />
            <Skeleton />
            <Skeleton />
          </Container>
        </Box>
      </Paper>
    </Box>
  );
}

DeploymentRequestsPage.pageTitle = 'Deployment requests';
