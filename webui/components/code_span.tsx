import Box from '@material-ui/core/Box';

export interface ICodeSpan {
  children: any;
}

export default function CodeSpan(props: ICodeSpan): JSX.Element {
  return (
    <Box component="code" color="info.dark">{ props.children }</Box>
  );
}
