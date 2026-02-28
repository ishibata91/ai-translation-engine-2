import React, { useState } from 'react';
import { NavLink } from 'react-router-dom';

const Sidebar: React.FC = () => {
    const [isCollapsed, setIsCollapsed] = useState(false);

    return (
        <div className={`flex flex-col h-full bg-base-200 transition-all duration-300 ${isCollapsed ? 'w-[72px]' : 'w-64'}`}>
            <div className={`flex items-center p-2 border-b border-base-300 ${isCollapsed ? 'justify-center' : 'justify-end'}`}>
                <button
                    className="btn btn-square btn-sm btn-ghost"
                    onClick={() => setIsCollapsed(!isCollapsed)}
                    title={isCollapsed ? "展開する" : "折りたたむ"}
                >
                    {isCollapsed ? (
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                            <path strokeLinecap="round" strokeLinejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" />
                        </svg>
                    ) : (
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
                        </svg>
                    )}
                </button>
            </div>

            <ul className="menu p-3 w-full space-y-2">
                <li title="ダッシュボード">
                    <NavLink to="/" className={({ isActive }) => (`flex items-center p-2 ${isActive ? 'active' : ''} ${isCollapsed ? 'justify-center' : ''}`)}>
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6 shrink-0"><path strokeLinecap="round" strokeLinejoin="round" d="m2.25 12 8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" /></svg>
                        {!isCollapsed && <span className="ml-1 font-medium truncate">ダッシュボード</span>}
                    </NavLink>
                </li>
                <li title="辞書構築">
                    <NavLink to="/dictionary" className={({ isActive }) => (`flex items-center p-2 ${isActive ? 'active' : ''} ${isCollapsed ? 'justify-center' : ''}`)}>
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6 shrink-0"><path strokeLinecap="round" strokeLinejoin="round" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25" /></svg>
                        {!isCollapsed && <span className="ml-1 font-medium truncate">辞書構築</span>}
                    </NavLink>
                </li>
                <li title="マスターペルソナ構築">
                    <NavLink to="/master_persona" className={({ isActive }) => (`flex items-center p-2 ${isActive ? 'active' : ''} ${isCollapsed ? 'justify-center' : ''}`)}>
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6 shrink-0"><path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" /></svg>
                        {!isCollapsed && <span className="ml-1 font-medium truncate">マスターペルソナ構築</span>}
                    </NavLink>
                </li>
                <li title="翻訳プロジェクト">
                    <NavLink to="/translation_flow" className={({ isActive }) => (`flex items-center p-2 ${isActive ? 'active' : ''} ${isCollapsed ? 'justify-center' : ''}`)}>
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6 shrink-0"><path strokeLinecap="round" strokeLinejoin="round" d="M12 21v-8.25M15.75 21v-8.25M8.25 21v-8.25M3 9l9-6 9 6m-1.5 12V10.332A48.36 48.36 0 0 0 12 9.75c-2.551 0-5.056.2-7.5.582V21M3 21h18M12 6.75h.008v.008H12V6.75Z" /></svg>
                        {!isCollapsed && <span className="ml-1 font-medium truncate">翻訳プロジェクト</span>}
                    </NavLink>
                </li>
                <li title="設定">
                    <NavLink to="/settings" className={({ isActive }) => (`flex items-center p-2 ${isActive ? 'active' : ''} ${isCollapsed ? 'justify-center' : ''}`)}>
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6 shrink-0"><path strokeLinecap="round" strokeLinejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.99l1.003.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.281Z" /><path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /></svg>
                        {!isCollapsed && <span className="ml-1 font-medium truncate">設定</span>}
                    </NavLink>
                </li>
            </ul>
        </div>
    );
};

export default Sidebar;
