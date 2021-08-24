import { useState, useEffect, useRef } from 'react';
import LinearProgress from '@material-ui/core/LinearProgress';

interface IProps {
  active: boolean;
  position: string | undefined;
}

// A spinner for indicating that data is being loaded. To avoid making the UI
// look too busy, it only shows up after a short timeout.
export default function DataLoadSpinner(props: IProps) {
  const { active, position } = props;
  const [show, setShow] = useState(false);
  const timerRef = useRef<number>();

  function getDivStyle(): object {
    var result: any = {
      position: 'relative',
    };
    if (!show) {
      result.visibility = 'hidden';
    }
    return result;
  }

  function getProgressStyle(): object {
    var result: any = {
      width: '100%',
    };
    if (position) {
      result.position = position;
    }
    return result;
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

  return (
    <div style={getDivStyle()}>
      <LinearProgress style={getProgressStyle()} />
    </div>
  );
}
