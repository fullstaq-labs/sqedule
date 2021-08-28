import { IAppContext } from '../components/app_context';
import { NavigationSection } from '../components/navbar';
import Container from '@material-ui/core/Container';
import Box from '@material-ui/core/Box';
import Typography from '@material-ui/core/Typography';
import MaterialLink from '@material-ui/core/Link';
import Button from '@material-ui/core/Button';
import imageStyles from '../common/images.module.css';

interface IProps {
  appContext: IAppContext;
}

export default function DashboardPage(_props: IProps) {
  return (
    <>
      <Container maxWidth="md">
        <Box px={2} py={2} textAlign="center">
          <Typography variant="body1" color="textSecondary">
            <p>
              <MaterialLink href="https://docs.sqedule.io/user_guide/" target="_blank" rel="noopener">
                <img src="logo-sqedule-horizontal.svg" alt="Sqedule logo" className={imageStyles.img_responsive} style={{ maxHeight: '150px' }} />
              </MaterialLink>
            </p>
            <p>
              Welcome to Sqedule, the application release &amp; auditing platform.
            </p>
            <p>
              <MaterialLink component={LinkButton} href="https://docs.sqedule.io/user_guide/" target="_blank" rel="noopener">
                Read the docs
              </MaterialLink>
            </p>
          </Typography>
        </Box>
      </Container>
    </>
  );
}

DashboardPage.navigationSection = NavigationSection.Dashboard;
DashboardPage.pageTitle = 'Home';

function LinkButton(props: any): JSX.Element {
  const { children, ...rest } = props;
  return (
    <Button color="primary" size="large" {...rest}>
      { children }
    </Button>
  );
}
