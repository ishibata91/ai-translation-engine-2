import React from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import LogViewer from './LogViewer';

const Layout: React.FC = () => {
    return (
        <div className="flex flex-row h-screen overflow-hidden bg-base-100 text-base-content">
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
                <LogViewer />
            </div>
        </div>
    );
};

export default Layout;
