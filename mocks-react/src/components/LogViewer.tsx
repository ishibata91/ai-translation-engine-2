import React, { useState, useEffect, useCallback } from 'react';

const LogViewer: React.FC = () => {
    const [width, setWidth] = useState(320);
    const [isResizing, setIsResizing] = useState(false);

    const handleMouseDown = useCallback((e: React.MouseEvent) => {
        e.preventDefault();
        setIsResizing(true);
    }, []);

    useEffect(() => {
        if (!isResizing) return;

        const handleMouseMove = (e: MouseEvent) => {
            // LogViewer is positioned on the right edge of the screen
            let newWidth = window.innerWidth - e.clientX;
            // Clamp width between 200px and 800px
            newWidth = Math.max(200, Math.min(newWidth, 800));
            setWidth(newWidth);
        };

        const handleMouseUp = () => {
            setIsResizing(false);
        };

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);

        // Prevent text selection while dragging
        document.body.style.userSelect = 'none';

        return () => {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            document.body.style.userSelect = '';
        };
    }, [isResizing]);

    return (
        <div
            style={{ width: `${width}px` }}
            className="group relative flex flex-col p-4 gap-4 h-full bg-base-300 border-l border-base-200 shadow-md transition-shadow"
        >
            {/* Drag Handle */}
            <div
                className="absolute left-0 top-0 bottom-0 w-1.5 cursor-col-resize hover:bg-primary z-20 transition-colors"
                onMouseDown={handleMouseDown}
            />

            <div className="flex justify-between items-center text-sm">
                <div className="badge badge-outline">Log Viewer</div>
                <div className="tabs tabs-boxed tabs-sm">
                    <a className="tab tab-active">All</a>
                    <a className="tab">TraceID</a>
                </div>
            </div>
            <div className="flex flex-col gap-2 overflow-y-auto w-full h-full text-xs">
                <div className="alert p-2 rounded flex-col items-start gap-1">
                    <span>[INFO] Parsing mod structure...</span>
                    <span className="badge badge-ghost badge-sm font-mono opacity-50 text-[0.6rem]">TraceID: 1a2b3c</span>
                </div>
                <div className="alert alert-error p-2 rounded flex-col items-start gap-1">
                    <span>[ERROR] Unmapped term found.</span>
                    <span className="badge badge-ghost badge-sm font-mono opacity-50 text-[0.6rem]">TraceID: 1a2b3d</span>
                </div>
            </div>
        </div>
    );
};

export default LogViewer;
