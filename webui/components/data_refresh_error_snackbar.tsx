import { useEffect, useState } from 'react';
import Snackbar from '@material-ui/core/Snackbar';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import CloseIcon from '@material-ui/icons/Close';
import { formatErrorMessage } from '../common/utils';

interface IProps {
  error: any;
  refreshing: boolean;
  onReload: () => void;
}

// A snackbar that presents the fact that data cannot be refreshed,
// and presents the option to reload or dismiss.
export default function DataRefreshErrorSnackbar(props: IProps): JSX.Element {
  const { error, refreshing, onReload } = props;
  const [show, setShow] = useState(false);

  useEffect(function() {
    setShow(error !== undefined && !refreshing);
  }, [error, refreshing]);

  function handleReload() {
    onReload();
  }

  function handleClose() {
    setShow(false);
  }

  return <Snackbar
    open={show}
    message={show && error && `Error refreshing data: ${formatErrorMessage(error)}`}
    action={
      <>
        <Button color="secondary" size="small" onClick={handleReload}>
          Reload
        </Button>
        <IconButton size="small" aria-label="Close" color="inherit" onClick={handleClose}>
          <CloseIcon fontSize="small" />
        </IconButton>
      </>
    }
    />
}
