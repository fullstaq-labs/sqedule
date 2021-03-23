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
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemAvatar from '@material-ui/core/ListItemAvatar';
import ListItemText from '@material-ui/core/ListItemText';
import Divider from '@material-ui/core/Divider';
import AccessTimeIcon from '@material-ui/icons/AccessTime';
import Container from '@material-ui/core/Container';
import { ColDef } from '@material-ui/data-grid';
import { formatStateString as formatReleaseStateString } from '../releases';
import styles from '../../common/tables.module.scss';

interface IProps {
  appContext: IAppContext;
}

export default function ApprovalRulesetPage(props: IProps) {
  const { appContext } = props;
  const theme = useTheme();
  const viewIsLarge = useMediaQuery(theme.breakpoints.up('md'));
  const [tabIndex, setTabIndex] = useState(0);
  const applicationsDataGridState = useDataGrid();
  const releasesDataGridState = useDataGrid();

  const router = useRouter();
  const id = router.query.id as string;
  const hasId = typeof id !== 'undefined';

  const { data, error, isValidating, mutate } =
    useSWR(hasId ?
      `/v1/approval-rulesets/${encodeURIComponent(id)}` :
      null);

  declarePageTitle(appContext, getPageTitle());
  declareValidatingFetchedData(appContext, isValidating);

  function getPageTitle() {
    if (data) {
      return `${data.display_name} (${id})`;
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

  return (
    <>
      {data &&
        <DataRefreshErrorSnackbar error={error} refreshing={isValidating} onReload={mutate} />
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
        <Tab label="Rules" id="rules" aria-controls="tab-panel-rules" />
        <Tab label="Applications" id="tab-applications" arial-controls="tab-panel-applications" />
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
            data={data}
            error={error}
            mutate={mutate}
            />
        </TabPanel>
        <TabPanel value={tabIndex} index={1} id="version-history">
        </TabPanel>
        <TabPanel value={tabIndex} index={2} id="rules">
          <RulesTabContents
            data={data}
            error={error}
            mutate={mutate}
            />
        </TabPanel>
        <TabPanel value={tabIndex} index={3} id="applications" style={{ flexGrow: 1 }}>
          <ApplicationsTabContents
            dataGridState={applicationsDataGridState}
            data={data}
            error={error}
            mutate={mutate}
            />
        </TabPanel>
        <TabPanel value={tabIndex} index={4} id="releases" style={{ flexGrow: 1 }}>
          <ReleasesTabContents
            applicationId={id}
            dataGridState={releasesDataGridState}
            data={data}
            error={error}
            mutate={mutate}
            />
        </TabPanel>
      </SwipeableViews>
    </>
  );
}

ApprovalRulesetPage.navigationSection = NavigationSection.ApprovalRulesets;
ApprovalRulesetPage.pageTitle = 'Approval ruleset';
ApprovalRulesetPage.hasBackButton = true;


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
                <Link href={`/approval-rulesets/${encodeURIComponent(data.id)}`}>
                  <a>{data.id}</a>
                </Link>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Display name</TableCell>
              <TableCell>
                <Link href={`/approval-rulesets/${encodeURIComponent(data.id)}`}>
                  <a>{data.display_name}</a>
                </Link>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Description</TableCell>
              <TableCell>{data.description || 'N/A'}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Latest version</TableCell>
              <TableCell>{data.major_version_number}.{data.minor_version_number}</TableCell>
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


interface IRulesTabContentsProps {
  data: any;
  error: any;
  mutate: () => void;
}

function RulesTabContents(props: IRulesTabContentsProps) {
  const { data } = props;

  if (data) {
    const rules = data.approval_rules.map(renderApprovalRule)
    return <Paper><List>{rules}</List></Paper>;
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

function renderApprovalRule(rule: any, index: number) {
  const title = <Typography variant="h6">{humanizeApprovalRuleType(rule.type)}</Typography>;
  return (
    <>
      {index > 0 && <Divider variant="inset" component="li" />}
      <ListItem alignItems="flex-start">
        <ListItemAvatar>
          <AccessTimeIcon style={{ fontSize: '3rem' }} />
        </ListItemAvatar>
        <ListItemText
          primary={title}
          secondary={renderApprovalRuleDetails(rule)}
          />
      </ListItem>
    </>
  );
}

function humanizeApprovalRuleType(type: string): string {
  switch (type) {
    case "http_api":
      return "HTTP API";
    case "schedule":
      return "Schedule";
    case "manual":
      return "Manual approval";
    default:
      return humanizeUnderscoreString(type);
  }
}

function renderApprovalRuleDetails(rule: any): JSX.Element {
  switch (rule.type) {
    case "http_api":
      return renderHttpApiApprovalRuleDetails(rule);
    case "schedule":
      return renderScheduleApprovalRuleDetails(rule);
    case "manual":
      return renderManualApprovalRuleDetails(rule);
    default:
      return <></>;
  }
}

function renderHttpApiApprovalRuleDetails(_rule: any) {
  // TODO
  return <></>;
}

function renderScheduleApprovalRuleDetails(rule: any) {
  return (
    <ul>
      {rule.begin_time &&
        <li>Begin time: {rule.begin_time}</li>
      }
      {rule.end_time &&
        <li>End time: {rule.end_time}</li>
      }
      {rule.days_of_week &&
        <li>Days of week: {rule.days_of_week}</li>
      }
      {rule.days_of_month &&
        <li>Days of month: {rule.days_of_month}</li>
      }
      {rule.months_of_year &&
        <li>Months of year: {rule.months_of_year}</li>
      }
    </ul>
  );
}

function renderManualApprovalRuleDetails(_rule: any) {
  // TODO
  return <></>;
}


const APPLICATION_APPROVAL_RULESET_BINDING_COLUMNS: ColDef[] = [
  {
    field: 'id',
    headerName: 'ID',
    width: 150,
    valueGetter: ({ row }) => row.application.id,
    renderCell: ({ row }) => (
      <Link href={`/applications/${encodeURIComponent(row.application.id)}`}>
        <a>{row.application.id}</a>
      </Link>
    ),
  },
  {
    field: 'display_name',
    headerName: 'Display name',
    width: 250,
    valueGetter: ({ row }) => row.application.display_name,
    renderCell: ({ row }) => (
      <Link href={`/applications/${encodeURIComponent(row.application.id)}`}>
        <a>{row.application.display_name}</a>
      </Link>
    ),
  },
  {
    field: 'latest_version',
    headerName: 'Latest version',
    width: 130,
    valueGetter: ({ row }) => `${row.application.major_version_number}.${row.application.minor_version_number}`,
  },
  {
    field: 'mode',
    headerName: 'Mode',
    width: 120,
    valueFormatter: ({ value }) => humanizeUnderscoreString(value as string),
  },
  {
    field: 'enabled',
    headerName: 'Enabled',
    width: 120,
    valueGetter: ({ row }) => row.application.enabled,
    valueFormatter: ({ value }) => (value as boolean) ? '✅' : '❌',
  },
  {
    field: 'review_state',
    headerName: 'Review state',
    width: 150,
    valueGetter: ({ row }) => row.application.review_state,
    valueFormatter: ({ value }) => formatReviewStateString(value as string),
  },
  {
    field: 'updated_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Updated at',
    valueGetter: ({ row }) => row.application.created_at,
    valueFormatter: ({ value }) => formatDateTimeString(value as string),
  },
];

interface IApplicationsTabContentsProps {
  dataGridState: DataGridRequestedState;
  data: any;
  error: any;
  mutate: () => void;
}

function ApplicationsTabContents(props: IApplicationsTabContentsProps) {
  const { dataGridState, data } = props;

  function addID(binding: any) {
    return { id: binding.application.id, ...binding };
  }

  if (data) {
    if (data.application_approval_ruleset_bindings.length == 0 && dataGridState.requestedPage == 1) {
      return (
        <Container maxWidth="md">
          <Box px={2} py={2} textAlign="center">
            <Typography variant="h5" color="textSecondary">
              There are no applications bound to this approval ruleset.
            </Typography>
          </Box>
        </Container>
      );
    }

    const rulesetBindings =
      paginateArray(data.application_approval_ruleset_bindings, dataGridState.requestedPage, dataGridState.requestedPageSize).
      map(addID);

    return (
      <Paper style={{ display: 'flex', flexGrow: 1 }}>
        <DataGrid
          rows={rulesetBindings}
          columns={APPLICATION_APPROVAL_RULESET_BINDING_COLUMNS}
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


const RELEASE_APPROVAL_RULESET_BINDING_COLUMNS: ColDef[] = [
  {
    field: 'id',
    type: 'number',
    headerName: 'ID',
    width: 100,
    valueGetter: ({ row }) => row.release.id,
    renderCell: ({ row }) => (
      <Box style={{ flexGrow: 1 }}> {/* Make the content properly align right */}
        <Link href={`/releases/${encodeURIComponent(row.release.application.id)}/${row.release.id}`}>
          <a>{row.release.id}</a>
        </Link>
      </Box>
    ),
  },
  {
    field: 'application',
    headerName: 'Application',
    width: 250,
    valueGetter: ({ row }) => row.release.application.display_name,
    renderCell: ({ row }) => (
      <Link href={`/applications/${encodeURIComponent(row.release.application.id)}`}>
        <a>{row.release.application.display_name}</a>
      </Link>
    ),
  },
  {
    field: 'mode',
    headerName: 'Mode',
    width: 120,
    valueFormatter: ({ value }) => humanizeUnderscoreString(value as string),
  },
  {
    field: 'state',
    headerName: 'State',
    width: 150,
    valueGetter: ({ row }) => row.release.state,
    valueFormatter: ({ value }) => formatReleaseStateString(value as string),
  },
  {
    field: 'created_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Created at',
    valueGetter: ({ row }) => row.release.created_at,
    valueFormatter: ({ value }) => formatDateTimeString(value as string),
  },
  {
    field: 'finalized_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Finalized at',
    valueGetter: ({ row }) => row.release.finalized_at,
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

  function addID(binding: any) {
    return { id: binding.release.id, ...binding };
  }

  if (data) {
    if (data.release_approval_ruleset_bindings.length == 0 && dataGridState.requestedPage == 1) {
      return (
        <Container maxWidth="md">
          <Box px={2} py={2} textAlign="center">
            <Typography variant="h5" color="textSecondary">
              There are no releases bound to this approval ruleset.
            </Typography>
          </Box>
        </Container>
      );
    }

    const bindings =
      paginateArray(data.release_approval_ruleset_bindings, dataGridState.requestedPage, dataGridState.requestedPageSize).
      map(addID);

    return (
      <Paper style={{ display: 'flex', flexGrow: 1 }}>
        <DataGrid
          rows={bindings}
          columns={RELEASE_APPROVAL_RULESET_BINDING_COLUMNS}
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
