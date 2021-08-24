import useSWR from 'swr';
import Link from 'next/link';
import { formatDateTimeString, formatProposalStateString, formatBooleanAsIcon } from '../common/utils';
import { IAppContext, declareValidatingFetchedData } from '../components/app_context';
import { NavigationSection } from '../components/navbar';
import DataRefreshErrorSnackbar from '../components/data_refresh_error_snackbar';
import DataLoadErrorScreen from '../components/data_load_error_screen';
import { DataGrid, useDataGrid } from '../components/data_grid';
import Container from '@material-ui/core/Container';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
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
      <Link href={`/approval-rulesets/${encodeURIComponent(row._orig_id)}`}>
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
      <Link href={`/approval-rulesets/${encodeURIComponent(row._orig_id)}`}>
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
    valueGetter: ({ row }) => row.latest_approved_version?.proposal_state,
    valueFormatter: ({ value }) => formatProposalStateString(value as any) ?? 'N/A',
  },
  {
    field: 'num_bound_applications',
    headerName: 'Applications bound',
    type: 'number',
    width: 160,
  },
  {
    field: 'num_bound_releases',
    headerName: 'Releases bound',
    type: 'number',
    width: 140,
    valueGetter: ({ row }) => row.latest_approved_version?.num_bound_releases,
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
    valueGetter: ({ row }) => row.latest_approved_version?.updated_at,
    valueFormatter: ({ value }) => formatDateTimeString(value as any) ?? 'N/A',
  },
];

export default function ApprovalRulesetsPage(props: IProps) {
  const { appContext } = props;
  const dataGridState = useDataGrid();
  const { data, error, isValidating, mutate } = useSWR(`/v1/approval-rulesets?page=${dataGridState.requestedPage}&per_page=${dataGridState.requestedPageSize}`);

  declareValidatingFetchedData(appContext, isValidating);

  if (data) {
    if (data.items.length == 0 && dataGridState.requestedPage == 1) {
      return (
        <Container maxWidth="md">
          <Box px={2} py={2} textAlign="center">
            <Typography variant="h5" color="textSecondary">
              There are no approval rulesets.
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

ApprovalRulesetsPage.navigationSection = NavigationSection.ApprovalRulesets;
ApprovalRulesetsPage.pageTitle = 'Approval rulesets';
