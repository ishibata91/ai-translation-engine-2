import React, { useState } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import LogViewer from './LogViewer';
import DetailPane from './DetailPane';
import LogDetail, { type LogEntry } from './LogDetail';

const Layout: React.FC = () => {
    const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null);

    return (
        <div className="flex flex-col h-screen overflow-hidden bg-base-100 text-base-content">
            {/* Top area: Sidebar + Main Content + LogViewer */}
            <div className="flex flex-row flex-1 overflow-hidden min-h-0">
                {/* Sidebar */}
                <div className="shrink-0 shadow-lg z-10 transition-all duration-300 relative">
                    <Sidebar />
                </div>

                {/* Main Content Area */}
                <div className="flex-1 flex flex-col relative overflow-hidden bg-base-300 p-4">
                    <div className="w-full h-full bg-base-100 rounded-xl shadow-sm overflow-y-auto">
                        <Outlet />
                    </div>
                </div>

                {/* Log Viewer */}
                <div className="shrink-0 z-10 shadow-[-4px_0_15px_-3px_rgba(0,0,0,0.1)]">
                    <LogViewer onLogClick={(log) => setSelectedLog(log)} />
                </div>
            </div>

            {/* Bottom area: Detail Pane for Logs */}
            <DetailPane
                isOpen={selectedLog !== null}
                onClose={() => setSelectedLog(null)}
                title={selectedLog ? `ログ詳細: ${selectedLog.level}` : 'ログ詳細'}
                defaultHeight={300}
            >
                <LogDetail log={selectedLog} />
            </DetailPane>
        </div>
    );
};

export default Layout;
