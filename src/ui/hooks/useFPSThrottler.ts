import { useState, useEffect, useRef } from 'react';

/**
 * useFPSThrottler
 * 
 * A hook that limits the frequency of state updates to a target FPS.
 * Essential for smooth terminal streaming without flickering.
 */
export function useFPSThrottler<T>(initialValue: T, fps = 15): [T, (val: T) => void] {
  const [displayValue, setDisplayValue] = useState<T>(initialValue);
  const pendingValue = useRef<T>(initialValue);
  const lastUpdate = useRef<number>(0);
  const interval = 1000 / fps;

  const updateValue = (val: T) => {
    pendingValue.current = val;
    const now = Date.now();
    
    if (now - lastUpdate.current >= interval) {
      setDisplayValue(val);
      lastUpdate.current = now;
    }
  };

  // Ensure the final value is always displayed
  useEffect(() => {
    const timer = setInterval(() => {
      if (displayValue !== pendingValue.current) {
        setDisplayValue(pendingValue.current);
      }
    }, interval);
    
    return () => clearInterval(timer);
  }, [displayValue, interval]);

  return [displayValue, updateValue];
}
