import { useState } from 'react';
import useSWR from 'swr';
import Link from 'next/link';
import { formatDateTimeString, formatProposalStateString, formatBooleanAsIcon } from '../common/utils';
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
import MaterialLink from '@material-ui/core/Link';
import Skeleton from '@material-ui/lab/Skeleton';
import { ColDef } from '@material-ui/data-grid';

interface IProps {
  appContext: IAppContext;
}

const COLUMNS: ColDef[] = [
  {
    field: 'id',
    headerName: 'ID',
    width: 120,
    renderCell: ({ row }) => (
      <Link href={`/applications/${encodeURIComponent(row._orig_id)}`}>
        <a>{row._orig_id}</a>
      </Link>
    ),
  },
  {
    field: 'display_name',
    headerName: 'Display name',
    width: 250,
    valueGetter: ({ row }) => row.latest_approved_version?.display_name ?? row._orig_id,
    renderCell: ({ row }) => (
      <Link href={`/applications/${encodeURIComponent(row._orig_id)}`}>
        <a>{row.latest_approved_version?.display_name ?? row._orig_id}</a>
      </Link>
    ),
  },
  {
    field: 'latest_version',
    headerName: 'Latest version',
    width: 130,
    valueGetter: ({ row }) => row.latest_approved_version?.version_number,
    valueFormatter: ({ value }) => value ?? 'N/A',
  },
  {
    field: 'enabled',
    headerName: 'Enabled',
    width: 110,
    valueGetter: ({ row }) => row.latest_approved_version?.enabled,
    valueFormatter: ({ value }) => formatBooleanAsIcon(value as any) ?? 'N/A',
  },
  {
    field: 'proposal_state',
    headerName: 'Proposal state',
    width: 150,
    valueGetter: ({row }) => row.latest_approved_version?.proposal_state,
    valueFormatter: ({ value }) => formatProposalStateString(value as any) || 'N/A',
  },
  {
    field: 'created_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Created at',
    valueFormatter: ({ value }) => formatDateTimeString(value as string),
  },
  {
    field: 'updated_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Updated at',
    valueGetter: ({row }) => row.latest_approved_version?.updated_at,
    valueFormatter: ({ value }) => formatDateTimeString(value as any) ?? 'N/A',
  },
];

export default function ApplicationsPage(props: IProps) {
  const { appContext } = props;
  const dataGridState = useDataGrid();
  const { data, error, isValidating, mutate } = useSWR(`/v1/applications?page=${dataGridState.requestedPage}&per_page=${dataGridState.requestedPageSize}`);
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
                <p>There are no applications.</p>
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

ApplicationsPage.navigationSection = NavigationSection.Applications;
ApplicationsPage.pageTitle = 'Applications';

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
      <DialogTitle id="creation-dialog-title">Create application</DialogTitle>
      <DialogContent id="creation-dialog-description">
        <Typography variant="body1">
          <p style={{ marginTop: 0 }}>
            <MaterialLink href="https://docs.sqedule.io/user_guide/tasks/install-cli/" target="_blank" rel="noopener">Install the CLI</MaterialLink>
            {' '}
            and
            {' '}
            <MaterialLink href="https://docs.sqedule.io/user_guide/tasks/initial-cli-setup/" target="_blank" rel="noopener">set it up</MaterialLink>. Then run:
          </p>
          <CodeBlock>
            sqedule application create \{"\n"}
            {' '} --id {'<STRING>'} \{"\n"}
            {' '} --display-name {'<STRING>'} \{"\n"}
            {' '} --proposal-state final
          </CodeBlock>
          <p style={{ marginBottom: 0 }}>
            To learn about all possible options, run <CodeSpan>sqedule application create --help</CodeSpan>
          </p>
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={props.onClose} color="primary">Close</Button>
      </DialogActions>
    </>
  );
}
