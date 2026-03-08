import { useEffect, useRef } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';

/**
 * Wails イベント購読を共通化する Hook。
 * handler の更新には追従しつつ、購読自体は eventName 単位で管理する。
 */
export function useWailsEvent<TPayload>(
    eventName: string,
    handler: (payload: TPayload) => void,
): void {
    const handlerRef = useRef(handler);

    useEffect(() => {
        handlerRef.current = handler;
    }, [handler]);

    useEffect(() => {
        const unsubscribe = EventsOn(eventName, (payload: TPayload) => {
            handlerRef.current(payload);
        });

        return () => {
            if (typeof unsubscribe === 'function') {
                unsubscribe();
            }
        };
    }, [eventName]);
}


