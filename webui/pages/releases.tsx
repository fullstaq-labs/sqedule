import { useState } from 'react';
import useSWR from 'swr';
import Link from 'next/link';
import { formatDateTimeString, humanizeUnderscoreString } from '../common/utils';
import { IAppContext, declareValidatingFetchedData } from '../components/app_context';
import { NavigationSection } from '../components/navbar';
import SecondaryToolbar from '../components/secondary_toolbar';
import CodeBlock from '../components/code_block';
import CodeSpan from '../components/code_span';
import DataRefreshErrorSnackbar from '../components/data_refresh_error_snackbar';
import DataLoadErrorScreen from '../components/data_load_error_screen';
import { DataGrid, useDataGrid } from '../components/data_grid';
import Container from '@material-ui/core/Container';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import AddCircleIcon from '@material-ui/icons/AddCircle';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';
import Skeleton from '@material-ui/lab/Skeleton';
import { ColDef } from '@material-ui/data-grid';

interface IProps {
  appContext: IAppContext;
}

const COLUMNS: ColDef[] = [
  {
    field: 'id',
    type: 'number',
    headerName: 'ID',
    width: 100,
    renderCell: ({ row }) => (
      <Box style={{ flexGrow: 1 }}> {/* Make the content properly align right */}
        <Link href={`/releases/${encodeURIComponent(row.application.id)}/${encodeURIComponent(row._orig_id)}`}>
          <a>{row._orig_id}</a>
        </Link>
      </Box>
    ),
  },
  {
    field: 'application_display_name',
    headerName: 'Application',
    width: 250,
    valueGetter: ({ row }) => row.application.latest_approved_version?.display_name ?? row.application.id,
    renderCell: ({ row }) => (
      <Link href={`/applications/${encodeURIComponent(row.application.id)}`}>
        <a>{row.application.latest_approved_version?.display_name ?? row.application.id}</a>
      </Link>
    ),
  },
  {
    field: 'state',
    headerName: 'State',
    width: 150,
    valueFormatter: ({ value }) => formatStateString(value as string),
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
    valueFormatter: ({ value }) => formatDateTimeString(value as any) ?? 'N/A',
  },
];

export default function ReleasesPage(props: IProps) {
  const { appContext } = props;
  const dataGridState = useDataGrid();
  const { data, error, isValidating, mutate } = useSWR(`/v1/releases?page=${dataGridState.requestedPage}&per_page=${dataGridState.requestedPageSize}`);
  const [creationDialogOpened, setCreationDialogOpened] = useState(false);

  function handleCreationDialogClose() {
    setCreationDialogOpened(false);
  }

  declareValidatingFetchedData(appContext, isValidating);

  if (data) {
    if (data.items.length == 0 && dataGridState.requestedPage == 1) {
      return (
        <>
          <Container maxWidth="md">
            <Box px={2} py={2} textAlign="center">
              <Typography variant="h5" color="textSecondary">
                <p>There are no releases.</p>
              </Typography>
              <Button size="large" variant="contained" color="primary" startIcon={ <AddCircleIcon/> } onClick={() => setCreationDialogOpened(true)}>Create</Button>
            </Box>
          </Container>

          <Dialog
            onClose={handleCreationDialogClose}
            aria-labelledby="creation-dialog-title"
            aria-describedby="creation-dialog-description"
            open={creationDialogOpened}
          >
            <CreationDialogContents onClose={handleCreationDialogClose} />
          </Dialog>
        </>
      );
    }

    return (
      <>
        <DataRefreshErrorSnackbar error={error} refreshing={isValidating} onReload={mutate} />
        <SecondaryToolbar><SecondaryToolbarContents/></SecondaryToolbar>
        <Box mx={2} my={2} style={{ display: 'flex', flexGrow: 1 }}>
          <Paper style={{ display: 'flex', flexGrow: 1 }}>
            <DataGrid
              rows={data.items}
              columns={COLUMNS}
              requestedState={dataGridState}
              style={{ flexGrow: 1 }} />
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

ReleasesPage.navigationSection = NavigationSection.Releases;
ReleasesPage.pageTitle = 'Releases';


export function formatStateString(state: string) {
  switch (state) {
    case 'in_progress':
      return '🕐\xa0 In progress';
    case 'cancelled':
      return '❕\xa0 Cancelled';
    case 'approved':
      return '✅\xa0 Approved';
    case 'rejected':
      return '❌\xa0 Rejected';
    default:
      return humanizeUnderscoreString(state);
  }
}

function SecondaryToolbarContents(): JSX.Element {
  const [creationDialogOpened, setCreationDialogOpened] = useState(false);

  function handleCreationDialogClose() {
    setCreationDialogOpened(false);
  }

  return (
    <>
      <Button color="primary" startIcon={ <AddCircleIcon/> } onClick={() => setCreationDialogOpened(true)}>Create</Button>

      <Dialog
        onClose={handleCreationDialogClose}
        aria-labelledby="creation-dialog-title"
        aria-describedby="creation-dialog-description"
        open={creationDialogOpened}
      >
        <CreationDialogContents onClose={handleCreationDialogClose} />
      </Dialog>
    </>
  );
}

function CreationDialogContents(props: any): JSX.Element {
  return (
    <>
      <DialogTitle id="creation-dialog-title">Create release</DialogTitle>
      <DialogContent id="creation-dialog-description">
        <Typography variant="body1">
          <p style={{ marginTop: 0 }}>
            <a href="https://docs.sqedule.io/user_guide/tasks/install-cli/" target="_blank">Install the CLI</a>
            {' '}
            and
            {' '}
            <a href="https://docs.sqedule.io/user_guide/tasks/initial-cli-setup/" target="_blank">set it up</a>. Then run:
          </p>
          <CodeBlock>
            sqedule release create --application-id {'<STRING>'}
          </CodeBlock>
          <p style={{ marginBottom: 0 }}>
            To learn about all possible options, run <CodeSpan>sqedule release create --help</CodeSpan>
          </p>
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={props.onClose} color="primary">Close</Button>
      </DialogActions>
    </>
  );
}
