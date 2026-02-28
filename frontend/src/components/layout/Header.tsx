import React from 'react';
import { useTheme } from '../../hooks/useTheme';

const Header: React.FC = () => {
    const { theme, setTheme, availableThemes } = useTheme();

    return (
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
                    {availableThemes.map((t) => (
                        <option key={t} value={t}>
                            {t.charAt(0).toUpperCase() + t.slice(1)}
                        </option>
                    ))}
                </select>
            </div>
        </header>
    );
};

export default Header;
