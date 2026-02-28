import React from 'react';
import { HashRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/layout/Layout';
import Dashboard from './pages/Dashboard';

const PlaceholderPage: React.FC<{ title: string }> = ({ title }) => (
    <div className="flex items-center justify-center h-full w-full">
        <h2 className="text-2xl font-bold opacity-50">{title} (Under Construction)</h2>
    </div>
);

function App() {
    return (
        <HashRouter>
            <Routes>
                <Route path="/" element={<Layout />}>
                    <Route index element={<Dashboard />} />
                    <Route path="dictionary" element={<PlaceholderPage title="辞書構築" />} />
                    <Route path="master_persona" element={<PlaceholderPage title="マスターペルソナ構築" />} />
                    <Route path="translation_flow" element={<PlaceholderPage title="翻訳プロジェクト" />} />
                    <Route path="settings" element={<PlaceholderPage title="設定" />} />
                    <Route path="theme" element={<PlaceholderPage title="テーマ確認" />} />
                </Route>
            </Routes>
        </HashRouter>
    );
}

export default App;
