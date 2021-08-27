import Image from 'next/image';
import { IAppContext } from '../components/app_context';
import { NavigationSection } from '../components/navbar';
import Container from '@material-ui/core/Container';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import MaterialLink from '@material-ui/core/Link';
import Button from '@material-ui/core/Button';

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
                <Image src="/logo-sqedule-horizontal.svg" alt="Sqedule logo" height={150} width={508} />
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
      {/* <Box mx={2} my={2} style={{ display: 'flex', flexGrow: 1 }}>
        <Paper style={{ flexGrow: 1 }}>
          <Box mx={2}>
            <Typography variant="body1">
              <p>
                <MaterialLink href="https://github.com/fullstaq-labs/sqedule" target="_blank" rel="noopener">
                  <Image src="/logo-sqedule-horizontal.svg" alt="Sqedule logo" height={150} width={508} />
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
        </Paper>
      </Box> */}
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
