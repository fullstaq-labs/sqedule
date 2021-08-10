import { useState } from 'react';
import { useRouter } from 'next/router';
import SwipeableViews from 'react-swipeable-views';
import Link from 'next/link';
import useSWR from 'swr';
import { formatStateString } from '../../releases';
import { formatDateTimeString, humanizeUnderscoreString, paginateArray, formatAdjustmentStateString, formatBooleanAsIcon } from '../../../common/utils';
import { IAppContext, declarePageTitle, declareValidatingFetchedData } from '../../../components/app_context';
import { NavigationSection } from '../../../components/navbar';
import DataRefreshErrorSnackbar from '../../../components/data_refresh_error_snackbar';
import DataLoadErrorScreen from '../../../components/data_load_error_screen';
import { DataGrid, useDataGrid, RequestedState as DataGridRequestedState } from '../../../components/data_grid';
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
import Button from '@material-ui/core/Button';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemAvatar from '@material-ui/core/ListItemAvatar';
import ListItemText from '@material-ui/core/ListItemText';
import Divider from '@material-ui/core/Divider';
import Badge from '@material-ui/core/Badge';
import AddCircleOutlineIcon from '@material-ui/icons/AddCircleOutline';
import CheckIcon from '@material-ui/icons/Check';
import AccessTimeIcon from '@material-ui/icons/AccessTime';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import Container from '@material-ui/core/Container';
import { ColDef } from '@material-ui/data-grid';
import styles from '../../../common/tables.module.scss';
import badgeStyles from '../../../common/badges.module.scss';
import eventStyles from '../../../components/approval_rule_processing_event.module.scss';
import { pathToApprovalRuleset } from '../../../common/paths';

interface IProps {
  appContext: IAppContext;
}

export default function ReleasePage(props: IProps) {
  const { appContext } = props;
  const theme = useTheme();
  const viewIsLarge = useMediaQuery(theme.breakpoints.up('md'));
  const [tabIndex, setTabIndex] = useState(0);
  const approvalRulesetsDataGridState = useDataGrid();

  const router = useRouter();
  const applicationId = router.query.application_id as string;
  const hasApplicationId = typeof applicationId !== 'undefined';
  const id = router.query.id as string;
  const hasId = typeof id !== 'undefined';

  const { data: appData, error: appError, isValidating: appDataIsValidating, mutate: appDataMutate } =
    useSWR(hasApplicationId ?
      `/v1/applications/${encodeURIComponent(applicationId)}` :
      null);
  const { data: releaseData, error: releaseError, isValidating: releaseDataIsValidating, mutate: releaseDataMutate } =
    useSWR((hasApplicationId && hasId) ?
      `/v1/applications/${encodeURIComponent(applicationId)}/releases/${encodeURIComponent(id)}` :
      null);
  const hasAllData = appData && releaseData;
  const firstError = appError || releaseError;
  const isValidating = appDataIsValidating || releaseDataIsValidating;

  declarePageTitle(appContext, getPageTitle());
  declareValidatingFetchedData(appContext, isValidating);

  function getPageTitle() {
    if (appData) {
      return `Release ${id} (for ${appData.latest_approved_version?.display_name ?? appData.id})`;
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
    releaseDataMutate();
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
        <Tab label="Approval rulesets" id="tab-approval-rulesets" arial-controls="tab-panel-approval-rulesets" />
        <Tab label="Events" id="tab-events" arial-controls="tab-panel-events" />
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
            appData={appData}
            appError={appError}
            appDataMutate={appDataMutate}
            releaseData={releaseData}
            releaseError={releaseError}
            releaseDataMutate={releaseDataMutate}
            />
        </TabPanel>
        <TabPanel value={tabIndex} index={1} id="approval-rulesets" style={{ flexGrow: 1 }}>
          <ApprovalRulesetsTabContents
            dataGridState={approvalRulesetsDataGridState}
            data={releaseData}
            error={releaseError}
            mutate={releaseDataMutate}
            />
        </TabPanel>
        <TabPanel value={tabIndex} index={1} id="events">
          <EventsTabContents
            data={releaseData}
            error={releaseError}
            mutate={releaseDataMutate}
            />
        </TabPanel>
      </SwipeableViews>
    </>
  );
}

ReleasePage.navigationSection = NavigationSection.Releases;
ReleasePage.pageTitle = 'Release';
ReleasePage.hasBackButton = true;


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
  appData: any;
  appError: any;
  appDataMutate: () => void;
  releaseData: any;
  releaseError: any;
  releaseDataMutate: () => void;
}

function GeneralTabContents(props: IGeneralTabContentsProps) {
  const { appData, releaseData } = props;
  const hasAllData = appData && releaseData;
  const firstError = props.appError || props.releaseError;

  function mutateAll() {
    props.appDataMutate();
    props.releaseDataMutate();
  }

  if (hasAllData) {
    return (
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
                  <a>{appData.latest_approved_version?.display_name ?? appData.id}</a>
                </Link>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">State</TableCell>
              <TableCell>{formatStateString(releaseData.state as string)}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Created at</TableCell>
              <TableCell>{formatDateTimeString(releaseData.created_at as string)}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell component="th" scope="row">Finalized at</TableCell>
              <TableCell>{formatDateTimeString(releaseData.finalized_at) ?? 'N/A'}</TableCell>
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
    );
  }

  if (firstError) {
    return <DataLoadErrorScreen error={firstError} onReload={mutateAll} />;
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
      <Link href={pathToApprovalRuleset(row.approval_ruleset)}>
        <a>{row.approval_ruleset.id}</a>
      </Link>
    ),
  },
  {
    field: 'display_name',
    headerName: 'Display name',
    width: 250,
    valueGetter: ({ row }) => row.approval_ruleset.display_name ?? row.approval_ruleset.id,
    renderCell: ({ row }) => (
      <Link href={pathToApprovalRuleset(row.approval_ruleset)}>
        <a>{row.approval_ruleset.latest_approved_version?.display_name ?? row.approval_ruleset.id}</a>
      </Link>
    ),
  },
  {
    field: 'version',
    headerName: 'Version',
    width: 120,
    valueGetter: ({ row }) => row.approval_ruleset.latest_approved_version?.version_number,
    valueFormatter: ({ value }) => value ?? 'N/A',
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
    valueGetter: ({ row }) => row.approval_ruleset.latest_approved_version?.enabled,
    valueFormatter: ({ value }) => formatBooleanAsIcon(value as any) ?? 'N/A',
  },
  {
    field: 'adjustment_state',
    headerName: 'Adjustment state',
    width: 150,
    valueGetter: ({ row }) => row.approval_ruleset.latest_approved_version?.adjustment_state,
    valueFormatter: ({ value }) => formatAdjustmentStateString(value as any) ?? 'N/A',
  },
  {
    field: 'ruleset_updated_at',
    type: 'dateTime',
    width: 180,
    headerName: 'Ruleset updated at',
    valueGetter: ({ row }) => row.approval_ruleset.latest_approved_version?.created_at,
    valueFormatter: ({ value }) => formatDateTimeString(value as any) ?? 'N/A',
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


function EventsTabContents(_props: any) {
  return (
    <Paper>
      <List>
        <ListItem alignItems="flex-start">
          <ListItemAvatar><AddCircleOutlineIcon style={{ fontSize: '2.8rem' }} /></ListItemAvatar>
          <ListItemText
            primary={<Typography variant="h6">Release created</Typography>}
            secondary="2021-03-29 11:00:01" />
        </ListItem>
        <Divider variant="inset" component="li" />
        <ListItem alignItems="flex-start">
          <ListItemAvatar><AccessTimeIcon style={{ fontSize: '2.8rem' }} /></ListItemAvatar>
          <ListItemText
            primary={<Typography variant="h6"><TextWithBadge text="Schedule approval rule processed" status="success" /></Typography>}
            secondary={
              <>
                <ul className={eventStyles.details_list}>
                  <li>Processed at: 2021-03-29 11:00:01</li>
                  <li>Ruleset: <Link href="/approval-rulesets/only%20afternoon">only afternoon</Link></li>
                </ul>
                <Button size="small" endIcon={<ExpandMoreIcon />}>View rule details</Button>
              </>
            } />
        </ListItem>
        <Divider variant="inset" component="li" />
        <ListItem alignItems="flex-start">
          <ListItemAvatar><CheckIcon style={{ fontSize: '2.8rem' }} /></ListItemAvatar>
          <ListItemText
            primary={<Typography variant="h6"><TextWithBadge text="Release finalized" status="success" /></Typography>}
            secondary="2021-03-29 11:00:02" />
        </ListItem>
      </List>
    </Paper>
  );
}

interface ITextWithBadgeProps {
  text: string;
  status: 'success';
}

function TextWithBadge(props) {
  const { text, status } = props;
  return (
    <>
      {text}
      {' '}
      <Badge
        badgeContent={status}
        classes={{ badge: `${badgeStyles.text_only} ${badgeStyles[status]}` }}
        />
    </>
  );
}
