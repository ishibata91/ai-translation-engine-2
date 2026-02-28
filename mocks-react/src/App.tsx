import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import DictionaryBuilder from './pages/DictionaryBuilder';
import MasterPersona from './pages/MasterPersona';
import TranslationFlow from './pages/TranslationFlow';
import Settings from './pages/Settings';
import ThemeShowcase from './pages/ThemeShowcase';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="dictionary" element={<DictionaryBuilder />} />
          <Route path="master_persona" element={<MasterPersona />} />
          <Route path="translation_flow" element={<TranslationFlow />} />
          <Route path="settings" element={<Settings />} />
          <Route path="theme" element={<ThemeShowcase />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};

export default App;
