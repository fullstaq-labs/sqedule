import Box from '@material-ui/core/Box';
import Divider from '@material-ui/core/Divider';

export interface ISecondaryToolbar {
  children: any;
}

// A secondary toolbar that appears right under the primary (blue) toolbar, for performing actions
// in the context of the main content area.
export default function SecondaryToolbar(props: ISecondaryToolbar): JSX.Element {
  return (
    <>
      <Box p={1.5} bgcolor="background.paper">
        {/* Ensure proper alignment with the first item in the navbar */}
        <div style={{ marginTop: '3px', marginBottom: '2px' }}>
          {props.children}
        </div>
      </Box>
      <Divider/>
    </>
  );
}
