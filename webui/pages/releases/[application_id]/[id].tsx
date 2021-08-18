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
import CancelIcon from '@material-ui/icons/Cancel';
import CloudIcon from '@material-ui/icons/Cloud';
import AccessTimeIcon from '@material-ui/icons/AccessTime';
import ThumbsUpDownIcon from '@material-ui/icons/ThumbsUpDown';
import GavelIcon from '@material-ui/icons/Gavel';
//import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
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
  const applicationID = router.query.application_id as string;
  const hasApplicationID = typeof applicationID !== 'undefined';
  const id = router.query.id as string;
  const hasID = typeof id !== 'undefined';

  const { data: appData, error: appError, isValidating: appDataIsValidating, mutate: appDataMutate } =
    useSWR(hasApplicationID ?
      `/v1/applications/${encodeURIComponent(applicationID)}` :
      null);
  const { data: releaseData, error: releaseError, isValidating: releaseDataIsValidating, mutate: releaseDataMutate } =
    useSWR((hasApplicationID && hasID) ?
      `/v1/applications/${encodeURIComponent(applicationID)}/releases/${encodeURIComponent(id)}` :
      null);
  const { data: eventsData, error: eventsError, isValidating: eventsDataIsValidating, mutate: eventsDataMutate } =
    useSWR(hasApplicationID && hasID ?
      `/v1/applications/${encodeURIComponent(applicationID)}/releases/${encodeURIComponent(id)}/events` :
      null);
  const hasAllData = appData && releaseData && eventsData;
  const firstError = appError || releaseError || eventsError;
  const isValidating = appDataIsValidating || releaseDataIsValidating || eventsDataIsValidating;

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
            data={eventsData}
            error={eventsError}
            mutate={eventsDataMutate}
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
              <TableCell component="th" scope="row">Metadata</TableCell>
              <TableCell>{renderMetadata(releaseData.metadata)}</TableCell>
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

function renderMetadata(metadata: object): string | JSX.Element {
  if (Object.keys(metadata).length == 0) {
    return 'N/A';
  } else {
    return <pre style={{ margin: 0 }}>{JSON.stringify(metadata, null, 4)}</pre>;
  }
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
  const { dataGridState, data, error, mutate } = props;

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

  if (error) {
    return <DataLoadErrorScreen error={error} onReload={mutate} />;
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


interface IEventsTabContentsProps {
  data: any;
  error: any;
  mutate: () => void;
}

function EventsTabContents(props: IEventsTabContentsProps) {
  const { data, error, mutate } = props;

  if (data) {
    var listItems: Array<JSX.Element> = [];
    data.items.forEach((event, i: number) => {
      var itemContent: JSX.Element | undefined;

      switch (event.type) {
      case 'created':
        itemContent = <ReleaseCreatedEvent event={event} />;
        break;
      case 'cancelled':
        itemContent = <ReleaseCancelledEvent event={event} />;
        break;
      case 'rule_processed':
        itemContent = <ReleaseRuleProcessedEvent event={event} />;
        break;
      }

      if (typeof itemContent !== 'undefined') {
        if (i > 0) {
          listItems.push(<Divider variant="inset" component="li" />);
        }
        listItems.push(<ListItem alignItems="flex-start">{itemContent as JSX.Element}</ListItem>);
      }
    });

    return (
      <Paper>
        <List>
          {listItems}
        </List>
      </Paper>
    );
  }

  if (error) {
    return <DataLoadErrorScreen error={error} onReload={mutate} />;
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

function ReleaseCreatedEvent(props: any): JSX.Element {
  return (
    <>
      <ListItemAvatar><AddCircleOutlineIcon style={{ fontSize: '2.8rem' }} /></ListItemAvatar>
      <ListItemText
        primary={<Typography variant="h6">Release created</Typography>}
        secondary={formatDateTimeString(props.event.created_at)} />
    </>
  );
}

function ReleaseCancelledEvent(props: any): JSX.Element {
  return (
    <>
      <ListItemAvatar><CancelIcon style={{ fontSize: '2.8rem' }} /></ListItemAvatar>
      <ListItemText
        primary={<Typography variant="h6"><TextWithBadge text="Release cancelled" badgeText="error" badgeType="error" /></Typography>}
        secondary={formatDateTimeString(props.event.created_at)} />
    </>
  );
}

function ReleaseRuleProcessedEvent(props: any): JSX.Element {
  const { event } = props;

  return (
    <>
      <ListItemAvatar><ReleaseRuleProcessedIcon eventType={event.type} /></ListItemAvatar>
      <ListItemText
        primary={<Typography variant="h6"><ReleaseRuleProcessedHeadline event={event} /></Typography>}
        secondary={
          <>
            <ul className={eventStyles.details_list}>
              <li>Processed at: {formatDateTimeString(event.created_at)}</li>
              {/* <li>Ruleset: <Link href="/approval-rulesets/only%20afternoon">only afternoon</Link></li> */}
            </ul>
            {/* <Button size="small" endIcon={<ExpandMoreIcon />}>View rule details</Button> */}
          </>
        } />
    </>
  );
}

function ReleaseRuleProcessedIcon(props: any): JSX.Element {
  var Icon: any;
  switch (props.eventType) {
    case 'http_api':
      Icon = CloudIcon;
      break;
    case 'schedule':
      Icon = AccessTimeIcon;
      break;
    case 'manual':
      Icon = ThumbsUpDownIcon;
      break;
    default:
      Icon = GavelIcon;
      break;
  }
  return (<Icon style={{ fontSize: '2.8rem' }} />);
}

function ReleaseRuleProcessedHeadline(props: any): JSX.Element {
  var typeName: string;
  var badgeType: BadgeType;

  switch (props.event.approval_rule_outcome.type) {
    case 'http_api':
      typeName = 'HTTP API';
      break;
    case 'schedule':
      typeName = 'Schedule';
      break;
    case 'manual':
      typeName = 'Manual';
      break;
    default:
      typeName = humanizeUnderscoreString(props.event.type) as string;
      break;
  }

  switch (props.event.result_state) {
  case 'in_progress':
    badgeType = 'neutral';
    break;
  case 'cancelled':
    badgeType = 'error';
    break;
  case 'approved':
    badgeType = 'success';
    break;
  case 'rejected':
    badgeType = 'error';
    break;
  default:
    badgeType = 'neutral';
    break;
  }

  return (
    <TextWithBadge
      text={`${typeName} approval rule processed`}
      badgeText={humanizeUnderscoreString(props.event.result_state) as string}
      badgeType={badgeType}
      />
  );
}


type BadgeType = 'neutral' | 'success' | 'error';

interface ITextWithBadgeProps {
  text: string;
  badgeText: string;
  badgeType: BadgeType;
}

function TextWithBadge(props: ITextWithBadgeProps): JSX.Element {
  return (
    <>
      {props.text}
      {' '}
      <Badge
        badgeContent={props.badgeText}
        classes={{ badge: `${badgeStyles.text_only} ${badgeStyles[props.badgeType]}` }}
        />
    </>
  );
}
