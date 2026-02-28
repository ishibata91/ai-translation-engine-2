import React, { useState, useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import LogViewer from './LogViewer';
import DetailPane from './DetailPane';
import LogDetail, { type LogEntry } from './LogDetail';

const THEMES = [
    "dark", "light", "cupcake", "bumblebee", "emerald",
    "corporate", "synthwave", "retro", "cyberpunk", "valentine"
];

const Layout: React.FC = () => {
    const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null);
    const [theme, setTheme] = useState(() => localStorage.getItem('theme') || 'dark');

    useEffect(() => {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
    }, [theme]);

    return (
        <div className="flex flex-col h-screen overflow-hidden bg-base-100 text-base-content transition-colors duration-200">
            {/* App Header */}
            <header className="navbar bg-base-200 border-b border-base-300 min-h-[3rem] h-12 shrink-0 z-20 shadow-sm">
                <div className="flex-1">
                    <a className="btn btn-ghost normal-case text-lg h-auto min-h-0 py-1 px-2 font-bold">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} className="w-6 h-6 stroke-primary"><path strokeLinecap="round" strokeLinejoin="round" d="m10.5 21 5.25-11.25L21 21m-9-3h7.5M3 5.621a48.474 48.474 0 0 1 6-.371m0 0c1.12 0 2.233.038 3.334.114M9 5.25V3m3.334 2.364C11.176 10.658 7.69 15.08 3 17.502m9.334-12.138c.896.061 1.785.147 2.666.257m-4.589 8.495a18.023 18.023 0 0 1-3.827-5.802" /></svg>
                        AI Translation Engine 2
                    </a>
                </div>
                <div className="flex flex-none items-center gap-2 px-2">
                    <span className="text-sm font-bold">Theme:</span>
                    <select
                        className="select select-bordered select-sm bg-base-100"
                        value={theme}
                        onChange={(e) => setTheme(e.target.value)}
                    >
                        {THEMES.map((t) => (
                            <option key={t} value={t}>
                                {t.charAt(0).toUpperCase() + t.slice(1)}
                            </option>
                        ))}
                    </select>
                </div>
            </header>

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
