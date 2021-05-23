import { useState } from 'react';
import { useRouter } from 'next/router';
import SwipeableViews from 'react-swipeable-views';
import Link from 'next/link';
import useSWR from 'swr';
import { formatDateTimeString, humanizeUnderscoreString, paginateArray, formatReviewStateString } from '../../common/utils';
import { IAppContext, declarePageTitle, declareValidatingFetchedData } from '../../components/app_context';
import { NavigationSection } from '../../components/navbar';
import DataRefreshErrorSnackbar from '../../components/data_refresh_error_snackbar';
import DataLoadErrorScreen from '../../components/data_load_error_screen';
import { DataGrid, useDataGrid, RequestedState as DataGridRequestedState } from '../../components/data_grid';
import { useTheme } from '@material-ui/core/styles';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import Box from '@material-ui/core/Box';
import Typography from '@material-ui/core/Typography';
import Paper from '@material-ui/core/Paper';
import Skeleton from '@material-ui/lab/Skeleton';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableRow from '@material-ui/core/TableRow';
import Container from '@material-ui/core/Container';
import { ColDef } from '@material-ui/data-grid';
import { formatStateString as formatReleaseStateString } from '../releases';
import styles from '../../common/tables.module.scss';

interface IProps {
  appContext: IAppContext;
}

export default function ApplicationPage(props: IProps) {
  const { appContext } = props;
  const theme = useTheme();
  const viewIsLarge = useMediaQuery(theme.breakpoints.up('md'));
  const [tabIndex, setTabIndex] = useState(0);
  const approvalRulesetsDataGridState = useDataGrid();
  const releasesDataGridState = useDataGrid();

  const router = useRouter();
  const id = router.query.id as string;
  const hasId = typeof id !== 'undefined';

  const { data: appData, error: appError, isValidating: appDataIsValidating, mutate: appDataMutate } =
    useSWR(hasId ?
      `/v1/applications/${encodeURIComponent(id)}` :
      null);
  const { data: releasesData, error: releasesError, isValidating: releasesDataIsValidating, mutate: releasesDataMutate } =
    useSWR(hasId ?
      `/v1/applications/${encodeURIComponent(id)}/releases?page=${releasesDataGridState.requestedPage}&per_page=${releasesDataGridState.requestedPageSize}` :
      null);
  const hasAllData = appData && releasesData;
  const firstError = appError || releasesError;
  const isValidating = appDataIsValidating || releasesDataIsValidating;

  declarePageTitle(appContext, getPageTitle());
  declareValidatingFetchedData(appContext, isValidating);

  function getPageTitle() {
    if (appData) {
      return `${appData.display_name} (${id})`;
    } else {
      return '';
    }
  }

  function handleTabChange(_event: React.ChangeEvent<{}>, newValue: number) {
    setTabIndex(newValue);
  }

  function handleTabIndexChange(index: number) {
    setTabIndex(index);
  };

  function mutateAll() {
    appDataMutate();
    releasesDataMutate();
  }

  return (
    <>
      {hasAllData &&
        <DataRefreshErrorSnackbar error={firstError} refreshing={isValidating} onReload={mutateAll} />
      }

      <Tabs
        value={tabIndex}
        onChange={handleTabChange}
        variant={viewIsLarge ? "standard" : "fullWidth"}
        indicatorColor="primary"
        textColor="primary"
      >
        <Tab label="General" id="tab-general" aria-controls="tab-panel-general" />
        <Tab label="Version history" id="version-history" aria-controls="tab-panel-version-history" />
        <Tab label="Approval rulesets" id="tab-approval-rulesets" arial-controls="tab-panel-approval-rulesets" />
        <Tab label="Releases" id="tab-releases" arial-controls="tab-panel-releases" />
      </Tabs>

      <SwipeableViews
        index={tabIndex}
        onChangeIndex={handleTabIndexChange}
        resistance={true}
        style={{ display: 'flex', flexDirection: 'column', flexGrow: 1 }}
        containerStyle={{ flexGrow: 1 }}
        slideStyle={{ display: 'flex', flexDirection: 'column' }}
      >
        <TabPanel value={tabIndex} index={0} id="general">
          <GeneralTabContents
            data={appData}
            error={appError}
            mutate={appDataMutate}
            />
        </TabPanel>
        <TabPanel value={tabIndex} index={1} id="version-history">
        </TabPanel>
        <TabPanel value={tabIndex} index={2} id="approval-rulesets" style={{ flexGrow: 1 }}>
          <ApprovalRulesetsTabContents
            dataGridState={approvalRulesetsDataGridState}
            data={appData}
            error={appError}
            mutate={appDataMutate}
            />
        </TabPanel>
        <TabPanel value={tabIndex} index={3} id="releases" style={{ flexGrow: 1 }}>
          <ReleasesTabContents
            applicationId={id}
            dataGridState={releasesDataGridState}
            data={releasesData}
            error={releasesError}
            mutate={releasesDataMutate}
            />
        </TabPanel>
      </SwipeableViews>
    </>
  );
}

ApplicationPage.navigationSection = NavigationSection.Applications;
ApplicationPage.pageTitle = 'Application';
ApplicationPage.hasBackButton = true;


interface ITabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
  id: string;
  style?: any;
}


function TabPanel(props: ITabPanelProps) {
  const { children, value, index, id, style, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`tab-panel-${id}`}
      aria-labelledby={`tab-${id}`}
      style={{ display: 'flex', ...style }}
      {...other}
    >
      <Box px={2} py={2} style={{ display: 'flex', flexDirection: 'column', flexGrow: 1 }}>
        {children}
      </Box>
    </div>
  );
}


interface IGeneralTabContentsProps {
  data: any;
  error: any;
  mutate: () => void;
}

function GeneralTabContents(props: IGeneralTabContentsProps) {
  const { data } = props;

  if (data) {
    return (
      <TableContainer component={Paper} className={styles.definition_list_table}>
        <Table>
          <TableBody>
            <TableRow>
              <TableCell component="th" scope="row">ID</TableCell>
              <TableCell>
                <Link href={`/applications/${encodeURIComponent(data.id)}`}>
                  <a>{data.id}</a>
                </Link>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Display name</TableCell>
              <TableCell>
                <Link href={`/applications/${encodeURIComponent(data.id)}`}>
                  <a>{data.display_name}</a>
                </Link>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Latest version</TableCell>
              <TableCell>{data.version_number}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Enabled</TableCell>
              <TableCell>{data.enabled ? '✅\xa0 Yes' : '❌\xa0 No'}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Review state</TableCell>
              <TableCell>{formatReviewStateString(data.review_state)}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Created at</TableCell>
              <TableCell>{formatDateTimeString(data.created_at as string)}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Updated at</TableCell>
              <TableCell>{formatDateTimeString(data.updated_at as string)}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    );
  }

  if (props.error) {
    return <DataLoadErrorScreen error={props.error} onReload={props.mutate} />;
  }

  return (
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
  );
}


const RULESET_BINDING_COLUMNS: ColDef[] = [
  {
    field: 'id',
    headerName: 'ID',
    width: 150,
    valueGetter: ({ row }) => row.approval_ruleset.id,
    renderCell: ({ row }) => (
      <Link href={`/approval-rulesets/${encodeURIComponent(row.approval_ruleset.id)}/versions/${encodeURIComponent(row.approval_ruleset.version_number)}`}>
        <a>{row.approval_ruleset.id}</a>
      </Link>
    ),
  },
  {
    field: 'display_name',
    headerName: 'Display name',
    width: 250,
    valueGetter: ({ row }) => row.approval_ruleset.display_name,
    renderCell: ({ row }) => (
      <Link href={`/approval-rulesets/${encodeURIComponent(row.approval_ruleset.id)}/versions/${encodeURIComponent(row.approval_ruleset.version_number)}`}>
        <a>{row.approval_ruleset.display_name}</a>
      </Link>
    ),
  },
  {
    field: 'latest_version',
    headerName: 'Latest version',
    width: 130,
    valueGetter: ({ row }) => row.approval_ruleset.version_number,
  },
  {
    field: 'enabled',
    headerName: 'Enabled',
    width: 120,
    valueGetter: ({ row }) => row.approval_ruleset.enabled,
    valueFormatter: ({ value }) => (value as boolean) ? '✅' : '❌',
  },
  {
    field: 'review_state',
    headerName: 'Review state',
    width: 150,
    valueGetter: ({ row }) => row.approval_ruleset.review_state,
    valueFormatter: ({ value }) => formatReviewStateString(value as string),
  },
  {
    field: 'mode',
    headerName: 'Mode',
    width: 120,
    valueFormatter: ({ value }) => humanizeUnderscoreString(value as string),
  },
  {
    field: 'updated_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Updated at',
    valueGetter: ({ row }) => row.approval_ruleset.created_at,
    valueFormatter: ({ value }) => formatDateTimeString(value as string),
  },
];

interface IApprovalRulesetsTabContentsProps {
  dataGridState: DataGridRequestedState;
  data: any;
  error: any;
  mutate: () => void;
}

function ApprovalRulesetsTabContents(props: IApprovalRulesetsTabContentsProps) {
  const { dataGridState, data } = props;

  function addID(rulesetBinding: any) {
    return { id: rulesetBinding.approval_ruleset.id, ...rulesetBinding };
  }

  if (data) {
    if (data.approval_ruleset_bindings.length == 0 && dataGridState.requestedPage == 1) {
      return (
        <Container maxWidth="md">
          <Box px={2} py={2} textAlign="center">
            <Typography variant="h5" color="textSecondary">
              There are no approval rulesets bound to this release.
            </Typography>
          </Box>
        </Container>
      );
    }

    const rulesetBindings =
      paginateArray(data.approval_ruleset_bindings, dataGridState.requestedPage, dataGridState.requestedPageSize).
      map(addID);

    return (
      <Paper style={{ display: 'flex', flexGrow: 1 }}>
        <DataGrid
          rows={rulesetBindings}
          columns={RULESET_BINDING_COLUMNS}
          requestedState={dataGridState}
          style={{ flexGrow: 1 }} />
      </Paper>
    );
  }

  if (props.error) {
    return <DataLoadErrorScreen error={props.error} onReload={props.mutate} />;
  }

  return (
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
  );
}


const RELEASE_COLUMNS: ColDef[] = [
  {
    field: 'id',
    type: 'number',
    headerName: 'ID',
    width: 120,
    renderCell: ({ row }) => (
      <Box style={{ flexGrow: 1 }}> {/* Make the content properly align right */}
        <Link href={`/releases/${encodeURIComponent(row.application_id)}/${encodeURIComponent(row._orig_id)}`}>
          <a>{row._orig_id}</a>
        </Link>
      </Box>
    ),
  },
  {
    field: 'state',
    headerName: 'State',
    width: 150,
    valueFormatter: ({ value }) => formatReleaseStateString(value as string),
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
    valueFormatter: ({ value }) => value ? formatDateTimeString(value as string) : 'N/A',
  },
];

interface IReleasesTabContentsProps {
  applicationId: string;
  dataGridState: DataGridRequestedState;
  data: any;
  error: any;
  mutate: () => void;
}

function ReleasesTabContents(props: IReleasesTabContentsProps) {
  const { dataGridState, data } = props;

  function addApplicationID(release: any) {
    return { application_id: props.applicationId, ...release };
  }

  if (data) {
    if (data.length == 0 && dataGridState.requestedPage == 1) {
      return (
        <Container maxWidth="md">
          <Box px={2} py={2} textAlign="center">
            <Typography variant="h5" color="textSecondary">
              There are no releases for this application.
            </Typography>
          </Box>
        </Container>
      );
    }

    const releases =
      paginateArray(data.items, dataGridState.requestedPage, dataGridState.requestedPageSize).
      map(addApplicationID);

    return (
      <Paper style={{ display: 'flex', flexGrow: 1 }}>
        <DataGrid
          rows={releases}
          columns={RELEASE_COLUMNS}
          requestedState={dataGridState}
          style={{ flexGrow: 1 }} />
      </Paper>
    );
  }

  if (props.error) {
    return <DataLoadErrorScreen error={props.error} onReload={props.mutate} />;
  }

  return (
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
  );
}
