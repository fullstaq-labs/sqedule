import React, { useState } from 'react';
import Hidden from '@material-ui/core/Hidden';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import ArrowBackIcon from '@material-ui/icons/ArrowBack';
import MenuIcon from '@material-ui/icons/Menu';
import { IUser } from '../common/user';
import DataLoadSpinner from './data_load_spinner';
import Navbar, { NavigationSection } from './navbar';
import styles from './layout.module.css';

interface IProps {
  navigationSection?: NavigationSection;
  title: string;
  hasBackButton: boolean;
  loading: boolean;
  user: IUser;
  children: any;
}

export default function Layout(props: IProps) {
  const { navigationSection, title, hasBackButton, loading, user, children } = props;
  const [navbarOpened, setNavbarOpened] = useState(false);

  function goBack() {
    history.back();
  }

  function openNavbar() {
    setNavbarOpened(true);
  }

  function closeNavbar() {
    setNavbarOpened(false);
  }

  return (
    <>
      <div className={styles.rootContainer}>
        <Hidden mdUp>
          {/* Mobile */}
          <Navbar variant="temporary" open={navbarOpened} navigationSection={navigationSection} user={user} showCloseButton={true} onCloseClicked={closeNavbar} />
        </Hidden>
        <Hidden smDown>
          {/* Desktop */}
          <Navbar variant="permanent" open={true} navigationSection={navigationSection} user={user} />
        </Hidden>

        <div className={styles.contentWrapper}>
          <AppBar position="relative">
            <Toolbar>
              {hasBackButton
                ? (
                  <IconButton edge="start" color="inherit" aria-label="Back" onClick={goBack}>
                    <ArrowBackIcon />
                  </IconButton>
                )
                : (
                  <Hidden mdUp>
                    <IconButton edge="start" color="inherit" aria-label="Menu" onClick={openNavbar}>
                      <MenuIcon />
                    </IconButton>
                  </Hidden>
                )
              }

              <Typography variant="h6" noWrap style={{flexGrow: 1}}>{title}</Typography>
              <Button color="inherit">Login</Button>
            </Toolbar>
          </AppBar>

          <DataLoadSpinner active={loading} />

          {children}
        </div>
      </div>
    </>
  )
}
