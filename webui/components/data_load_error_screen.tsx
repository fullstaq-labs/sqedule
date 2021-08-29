import grey from '@material-ui/core/colors/grey';
import Container from '@material-ui/core/Container';
import Box from '@material-ui/core/Box';
import Typography from '@material-ui/core/Typography';
import CloudOffIcon from '@material-ui/icons/CloudOff';
import Button from '@material-ui/core/Button';
import { formatErrorMessage } from '../common/utils';

interface IProps {
  error: any;
  onReload: () => void;
}

// An error screen that presents the fact that data cannot be loaded,
// and presents the option to reload.
export default function DataLoadErrorScreen(props: IProps): JSX.Element {
  const { error, onReload } = props;

  function handleReload() {
    onReload();
  }

  return (
    <Container maxWidth="md">
      <Box px={2} py={2} textAlign="center">
        <Typography color="textSecondary" paragraph={true}>
          <CloudOffIcon style={{ fontSize: '15rem', color: grey[400] }} />
        </Typography>
        <Typography variant="h5" color="textSecondary" paragraph={true}>
          Oops, something went wrong
        </Typography>
        <Typography paragraph={true}>
          {formatErrorMessage(error)}
        </Typography>
        <Button variant="contained" color="primary" onClick={handleReload}>Reload</Button>
      </Box>
    </Container>
  );
}
