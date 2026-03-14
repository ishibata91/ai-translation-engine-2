import React, {useEffect} from 'react';
import {HashRouter, Route, Routes} from 'react-router-dom';
import Layout from './components/layout/Layout';
import Dashboard from './pages/Dashboard';
import DictionaryBuilder from './pages/DictionaryBuilder';
import MasterPersona from './pages/MasterPersona';
import TranslationFlow from './pages/TranslationFlow';
import {initTaskListeners, useTaskStore} from './store/taskStore';

const PlaceholderPage: React.FC<{ title: string }> = ({ title }) => (
    <div className="flex items-center justify-center h-full w-full">
        <h2 className="text-2xl font-bold opacity-50">{title} (Under Construction)</h2>
    </div>
);

function App() {
    const fetchActiveTasks = useTaskStore(state => state.fetchActiveTasks);

    useEffect(() => {
        initTaskListeners();
        fetchActiveTasks();
    }, [fetchActiveTasks]);

    return (
        <HashRouter>
            <Routes>
                <Route path="/" element={<Layout />}>
                    <Route index element={<Dashboard />} />
                    <Route path="dictionary" element={<DictionaryBuilder />} />
                    <Route path="master_persona" element={<MasterPersona />} />
                    <Route path="translation_flow" element={<TranslationFlow />} />
                    <Route path="settings" element={<PlaceholderPage title="設定" />} />
                    <Route path="theme" element={<PlaceholderPage title="テーマ確認" />} />
                </Route>
            </Routes>
        </HashRouter>
    );
}

export default App;
