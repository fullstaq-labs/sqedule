import { useState, useEffect, useRef } from 'react';
import LinearProgress from '@material-ui/core/LinearProgress';

interface IProps {
  active: boolean;
}

// A spinner for indicating that data is being loaded. To avoid making the UI
// look too busy, it only shows up after a short timeout.
export default function DataLoadSpinner(props: IProps) {
  const { active } = props;
  const [show, setShow] = useState(false);
  const timerRef = useRef<number>();

  function getStyle(): object {
    if (show) {
      return {};
    } else {
      return { visibility: 'hidden' };
    }
  }

  useEffect(function() {
    if (active) {
      timerRef.current = window.setTimeout(function() {
        timerRef.current = undefined;
        setShow(true);
      }, 1000);

      return function() {
        if (timerRef.current !== null) {
          clearTimeout(timerRef.current);
          timerRef.current = undefined;
        }
      }
    } else {
      setShow(false);
    }
  }, [active]);

  return <LinearProgress style={getStyle()} />;
}
