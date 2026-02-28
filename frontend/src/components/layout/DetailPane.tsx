import React, { useState, useRef, useCallback, useEffect } from 'react';
import { useUIStore } from '../../store/uiStore';
import LogDetail from '../log/LogDetail';

const DetailPane: React.FC = () => {
    const { detailPane, closeDetailPane } = useUIStore();
    const { isOpen, type, payload } = detailPane;

    const defaultHeight = 300;
    const minHeight = 100;
    const maxHeight = 800;

    const [height, setHeight] = useState(defaultHeight);
    const isDragging = useRef(false);
    const startY = useRef(0);
    const startHeight = useRef(0);

    const onMouseMove = useCallback((e: MouseEvent) => {
        if (!isDragging.current) return;
        const delta = startY.current - e.clientY;
        const next = Math.min(maxHeight, Math.max(minHeight, startHeight.current + delta));
        setHeight(next);
    }, [minHeight, maxHeight]);

    const onMouseUp = useCallback(() => {
        if (!isDragging.current) return;
        isDragging.current = false;
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
    }, []);

    useEffect(() => {
        document.addEventListener('mousemove', onMouseMove);
        document.addEventListener('mouseup', onMouseUp);
        return () => {
            document.removeEventListener('mousemove', onMouseMove);
            document.removeEventListener('mouseup', onMouseUp);
        };
    }, [onMouseMove, onMouseUp]);

    const handleDragStart = (e: React.MouseEvent) => {
        isDragging.current = true;
        startY.current = e.clientY;
        startHeight.current = height;
        document.body.style.cursor = 'ns-resize';
        document.body.style.userSelect = 'none';
        e.preventDefault();
    };

    let title = '詳細';
    if (type === 'log' && payload) {
        title = `ログ詳細: ${payload.level}`;
    }

    return (
        <div
            className="w-full bg-base-100 border-t-2 border-primary flex flex-col shrink-0 overflow-hidden"
            style={{
                height: isOpen ? `${height}px` : 0,
                transition: isDragging.current ? 'none' : 'height 0.3s ease-in-out',
            }}
        >
            <div
                className="w-full flex justify-center items-center shrink-0 group"
                style={{ height: '8px', cursor: 'ns-resize' }}
                onMouseDown={handleDragStart}
                title="ドラッグして高さを調整"
            >
                <div className="w-12 h-1 rounded-full bg-base-300 group-hover:bg-primary transition-colors" />
            </div>

            <div className="flex justify-between items-center px-4 py-2 bg-base-200 border-b border-base-300 shrink-0">
                <div className="flex items-center gap-2">
                    <span className="w-1 h-4 bg-primary rounded-full inline-block" />
                    <span className="font-bold text-sm">{title}</span>
                </div>
                <div className="flex items-center gap-1">
                    <button
                        className="btn btn-ghost btn-xs"
                        onClick={() => setHeight(defaultHeight)}
                        title="デフォルト高さに戻す"
                    >
                        ⊟
                    </button>
                    <button
                        className="btn btn-ghost btn-xs"
                        onClick={closeDetailPane}
                        aria-label="詳細パネルを閉じる"
                        title="閉じる"
                    >
                        ✕
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 min-h-0">
                {type === 'log' && payload && <LogDetail log={payload} />}
            </div>
        </div>
    );
};

export default DetailPane;
