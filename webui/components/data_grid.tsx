import { CSSProperties, Dispatch, SetStateAction, useState } from 'react';
import { DataGrid as MaterialDataGrid, RowModel, Columns, SortModelParams } from '@material-ui/data-grid';
import Box from '@material-ui/core/Box';
import IconButton from '@material-ui/core/IconButton';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import { useIsomorphicLayoutEffect } from '../common/utils';
import styles from './data_grid.module.scss';

interface RequestedState {
  requestedPage: number;
  setRequestedPage: Dispatch<SetStateAction<number>>;
  requestedPageSize: number;
}

interface IProps {
  rows: RowModel[];
  columns: Columns;
  requestedState: RequestedState;
  style?: CSSProperties;
  onSortModelChange?: (params: SortModelParams) => void;
}

export const DEFAULT_PAGE_SIZE = 100;

export function useDataGrid(): RequestedState {
  const [requestedPage, setRequestedPage] = useState(1);
  return {
    requestedPage,
    setRequestedPage,
    requestedPageSize: DEFAULT_PAGE_SIZE,
  };
}

/**
 * A grid for displaying tabular data. This is the same as Material UI's
 * DataGrid, but with the following modifications suited to our use case:
 *
 *  - Always assume server-side pagination.
 *  - Don't require providing the total amount of rows. Allow "infinite paging".
 *  - Avoid infinite loops in the page change callback.
 *
 * ## Usage
 *
 * ~~~tsx
 * import { DataGrid, useDataGrid } from '../components/data_grid';
 *
 * function MyComponent() {
 *   const dataGridState = useDataGrid();
 *   const { data, error, isValidating, mutate } = useSWR(`/v1/api-call?page=${dataGridState.requestedPage}&per_page=${dataGridState.requestedPageSize}`);
 *
 *   // ...
 *
 *   return (
 *     <DataGrid
 *       requestedState={dataGridState}
 *
 *       // ...and other supported props:
 *       rows={data.items}
 *       columns={...}
 *       />
 *   );
 * }
 * ~~~
 *
 * ## How Material UI's DataGrid's page change callback is susceptible to infinite loops
 *
 * Sqedule's pages render a skeleton instead of DataGrid while data is being loaded.
 * This means that the DataGrid may be temporarily unmounted at any time, losing its
 * state (such as which page it's on). This means that DataGrid state which should not be
 * lost during the lifetime of a page, should live in the page component, and passed
 * to the DataGrid.
 *
 * This is roughly how the above is implemented using Material UI's DataGrid:
 *
 * ~~~tsx
 * import { MaterialDataGrid as DataGrid } from '@material-ui/data-grid';
 *
 * function MyComponent() {
 *   const [requestedPage, setRequestedPage] = useState(1);
 *
 *   function handlePageChange(params) {
 *     setRequtestedPage(params.page);
 *   }
 *
 *   return (
 *     <MaterialDataGrid
 *       page={requestedPage}
 *       onPageChange={handlePageChange}
 *       // ...
 *       />
 *   );
 * }
 * ~~~
 *
 * However, passing the `page` property may result in firing of the `onPageChange`
 * callback. That callback then rerenders MaterialDataGrid, which refires the callback.
 */
export function DataGrid(props: IProps) {
  const { rows, columns, requestedState, onSortModelChange, style } = props;
  const { requestedPage, setRequestedPage } = requestedState;
  const [page, setPage] = useState(requestedPage);

  function isPrevDisabled() {
    return page <= 1;
  }

  function isNextDisabled() {
    return rows.length == 0;
  }

  function handlePrevPage() {
    if (page > 1) {
      setPage(page - 1);
    }
    setRequestedPage(page - 1);
  }

  function handleNextPage() {
    setPage(page + 1);
    setRequestedPage(page + 1);
  }

  useIsomorphicLayoutEffect(function() {
    if (requestedPage !== undefined && requestedPage != page) {
      setPage(requestedPage);
    }
  }, [requestedPage, page, setPage]);

  return (
    <Box className={styles.outer_container} style={{ ...style }}>
      <Box className={styles.material_grid_container}>
        <MaterialDataGrid
          rows={rows}
          columns={columns}
          sortingMode="server"
          page={page}
          pageSize={100}
          hideFooter={true}
          onSortModelChange={onSortModelChange}
          />
      </Box>

      <Box p={0.5} textAlign="right">
        <span className={styles.page_label}>Page {page}</span>
        <IconButton aria-label="Previous" disabled={isPrevDisabled()} onClick={handlePrevPage}>
          <ChevronLeftIcon />
          </IconButton>
        <IconButton aria-label="Next" disabled={isNextDisabled()} onClick={handleNextPage}>
          <ChevronRightIcon />
          </IconButton>
      </Box>
    </Box>
  );
}
