import React, { useEffect } from 'react';
import { useRouter, NextRouter } from 'next/router';
import Link from 'next/link';
import Image from 'next/image';
import Drawer from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import Divider from '@material-ui/core/Divider';
import IconButton from '@material-ui/core/IconButton';
import MaterialLink from '@material-ui/core/Link';
import CloseIcon from '@material-ui/icons/Close';
import AccessAlarmIcon from '@material-ui/icons/AccessAlarm';
import HomeIcon from '@material-ui/icons/Home';
import AppsIcon from '@material-ui/icons/Apps';
import AssignmentIcon from '@material-ui/icons/Assignment';
import GavelIcon from '@material-ui/icons/Gavel';
import ThumbsUpDownIcon from '@material-ui/icons/ThumbsUpDown';
import BusinessIcon from '@material-ui/icons/Business';
import AccountBoxIcon from '@material-ui/icons/AccountBox';
import SettingsIcon from '@material-ui/icons/Settings';
import ExitToAppIcon from '@material-ui/icons/ExitToApp';
import HelpOutlineIcon from '@material-ui/icons/HelpOutline';
import InfoIcon from '@material-ui/icons/Info';
import { IUser } from '../common/user';
import styles from './navbar.module.scss';

export enum NavigationSection {
  Dashboard = 'dashboard',
  Applications = 'applications',
  Releases = 'releases',
  ApprovalRulesets = 'approval-rulesets',
  ManualApprovals = 'manual-approvals',
  About = 'about',
}

interface IProps {
  open: boolean;
  variant?: "permanent" | "persistent" | "temporary" | undefined;
  navigationSection?: NavigationSection;
  user: IUser;
  showCloseButton?: boolean;
  onCloseClicked?: () => void;
}

function useEffect_CloseNavbarOnRouteChange(router: NextRouter, onCloseClicked?: () => void) {
  useEffect(function() {
    function handleRouteChange() {
      if (onCloseClicked) {
        onCloseClicked();
      }
    }

    router.events.on('routeChangeStart', handleRouteChange);
    return function() {
      router.events.off('routeChangeStart', handleRouteChange);
    }
  }, []);
}

export default function Navbar(props: IProps) {
  const { open, variant, navigationSection, user, showCloseButton, onCloseClicked } = props;
  const router = useRouter();

  useEffect_CloseNavbarOnRouteChange(router, onCloseClicked);

  return (
    <Drawer variant={variant} open={open} className={styles.navbar} classes={{paper: styles.paper}} onClose={onCloseClicked}>
      <List classes={{ root: styles.app_banner }}>
        <ListItem>
          <MaterialLink href="https://github.com/fullstaq-labs/sqedule" underline="none" target="_blank" rel="noopener">
            <ListItemIcon>
              <Image src="/logo-sqedule-horizontal.svg" width={132} height={39} alt="Sqedule logo" />
            </ListItemIcon>
          </MaterialLink>
          {showCloseButton &&
            <ListItemSecondaryAction>
              <IconButton edge="end" aria-label="Close menu" onClick={onCloseClicked}>
                <CloseIcon />
              </IconButton>
            </ListItemSecondaryAction>
          }
        </ListItem>
      </List>
      <Divider />

      <List component="nav">
        {/* <Link href="/user">
          <ListItem button>
            <ListItemIcon><AccountBoxIcon /></ListItemIcon>
            <ListItemText primary={`${user.full_name}'s profile`} />
          </ListItem>
        </Link> */}
        <Link href="/">
          <ListItem button selected={navigationSection == NavigationSection.Dashboard}>
            <ListItemIcon><HomeIcon /></ListItemIcon>
            <ListItemText primary="Dashboard" />
          </ListItem>
        </Link>
      </List>

      <Divider />

      <List component="nav">
        <Link href="/applications">
          <ListItem button selected={navigationSection == NavigationSection.Applications}>
            <ListItemIcon><AppsIcon /></ListItemIcon>
            <ListItemText primary="Applications" />
          </ListItem>
        </Link>
        <Link href="/releases">
          <ListItem button selected={navigationSection == NavigationSection.Releases}>
            <ListItemIcon><AssignmentIcon /></ListItemIcon>
            <ListItemText primary="Releases" />
          </ListItem>
        </Link>
        <Link href="/approval-rulesets">
          <ListItem button selected={navigationSection == NavigationSection.ApprovalRulesets}>
            <ListItemIcon><GavelIcon /></ListItemIcon>
            <ListItemText primary="Approval rulesets" />
          </ListItem>
        </Link>
        {/* <Link href="/manual-approvals">
          <ListItem button selected={navigationSection == NavigationSection.ManualApprovals}>
            <ListItemIcon><ThumbsUpDownIcon /></ListItemIcon>
            <ListItemText primary="Manual approvals" />
          </ListItem>
        </Link> */}
      </List>

      <Divider />

      <List component="nav">
        {/* <Link href="/organization">
          <ListItem button>
            <ListItemIcon><BusinessIcon /></ListItemIcon>
            <ListItemText primary="Organization" />
          </ListItem>
        </Link>
        <Link href="/user">
          <ListItem button>
            <ListItemIcon><SettingsIcon /></ListItemIcon>
            <ListItemText primary="Settings" />
          </ListItem>
        </Link>
        <Link href="/user">
          <ListItem button>
            <ListItemIcon><ExitToAppIcon /></ListItemIcon>
            <ListItemText primary="Logout" />
          </ListItem>
        </Link> */}
        <MaterialLink href="https://docs.sqedule.io/user_guide/" color="textPrimary" underline="none" target="_blank" rel="noopener">
          <ListItem button>
            <ListItemIcon><HelpOutlineIcon /></ListItemIcon>
            <ListItemText primary="Help" />
          </ListItem>
        </MaterialLink>
        <Link href="/about">
          <ListItem button selected={navigationSection == NavigationSection.About}>
            <ListItemIcon><InfoIcon /></ListItemIcon>
            <ListItemText primary="About" />
          </ListItem>
        </Link>
      </List>
    </Drawer>
  )
}
