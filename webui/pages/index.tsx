import { IAppContext } from '../components/app_context';
import { NavigationSection } from '../components/navbar';
import Button from '@material-ui/core/Button';

interface IProps {
  appContext: IAppContext;
}

export default function DashboardPage(_props: IProps) {
  return <Button color="primary">hello world</Button>;
}

DashboardPage.navigationSection = NavigationSection.Dashboard;
DashboardPage.pageTitle = 'Home';
