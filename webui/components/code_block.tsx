import Box from '@material-ui/core/Box';

export interface ICodeBlock {
  children: any;
}

export default function CodeBlock(props: ICodeBlock): JSX.Element {
  return (
    <Box p={2} component="pre" color="primary.contrastText" bgcolor="grey.A400" style={{ overflow: 'auto' }}>
      {props.children}
    </Box>
  );
}
