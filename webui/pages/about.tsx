import useSWR from 'swr';
import { NavigationSection } from '../components/navbar';
import { IAppContext, declareValidatingFetchedData } from '../components/app_context';
import DataRefreshErrorSnackbar from '../components/data_refresh_error_snackbar';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import MaterialLink from '@material-ui/core/Link';
import Button from '@material-ui/core/Button';
import imageStyles from '../common/images.module.css';

interface IProps {
  appContext: IAppContext;
}

export default function AboutPage(props: IProps): JSX.Element {
  const { appContext } = props;
  const { data, error, isValidating, mutate } = useSWR(`/v1/about`);

  declareValidatingFetchedData(appContext, isValidating);

  var version: string;
  if (data) {
    version = data.version;
  } else if (error) {
    version = '<error loading>';
  } else {
    version = 'loading...';
  }

  return (
    <>
      {(error || isValidating) && <DataRefreshErrorSnackbar error={error} refreshing={isValidating} onReload={mutate} />}
      <Box mx={2} my={2} style={{ display: 'flex', flexGrow: 1 }}>
        <Paper style={{ flexGrow: 1 }}>
          <Box mx={2}>
            <Typography variant="body1">
              <h2>About Sqedule</h2>
              <ul>
                <li><strong>Server version</strong> — {version}</li>
                <li><strong>Github</strong> — <MaterialLink href="https://github.com/fullstaq-labs/sqedule" target="_blank" rel="noopener">fullstaq-labs/sqedule</MaterialLink></li>
              </ul>

              <h2>About Fullstaq</h2>
              <p>
                <MaterialLink href="https://fullstaq.com" target="_blank" rel="noopener">
                  <img src="../logo-fullstaq.svg" alt="Fullstaq logo" className={imageStyles.img_responsive} style={{ maxHeight: '99px' }} />
                </MaterialLink>
              </p>
              <p>
                Sqedule is made with ❤️&nbsp; by <MaterialLink href="https://fullstaq.com" target="_blank" rel="noopener">Fullstaq</MaterialLink>.
                {' '}
                Based in the Netherlands, Fullstaq helps organizations with complex IT environments to solve complex problems and to succeed in cloud native.
              </p>
              <p>
                Fullstaq helps by providing consultancy, training, managed services and project development services in the area of cloud native technologies, Kubernetes, containerization and DevOps.
              </p>
              <p>
                We differentiate ourselves with <strong>quality</strong> and <strong>passion</strong>. We've already helped clients such as ASML, Achmea, NS (Dutch Railways), Albert Heijn and Topgeschenken. What can we do for you?
              </p>
              <p>
                <MaterialLink component={LinkButton} href="https://fullstaq.com" target="_blank" rel="noopener">
                  Come and say hi
                </MaterialLink>
              </p>
            </Typography>
          </Box>
        </Paper>
      </Box>
    </>
  );
}

AboutPage.navigationSection = NavigationSection.About;
AboutPage.pageTitle = 'About';

function LinkButton(props: any): JSX.Element {
  const { children, ...rest } = props;
  return (
    <Button variant="contained" color="primary" size="large" {...rest}>
      { children }
    </Button>
  );
}
