import React, { useState } from 'react';
import { TerminologyPanel } from '../components/translation-flow/TerminologyPanel';
import { PersonaPanel } from '../components/translation-flow/PersonaPanel';
import { SummaryPanel } from '../components/translation-flow/SummaryPanel';
import { TranslationPanel } from '../components/translation-flow/TranslationPanel';
import { ExportPanel } from '../components/translation-flow/ExportPanel';

const TranslationFlow: React.FC = () => {
    const [activeTab, setActiveTab] = useState(0);

    const tabs = [
        { label: '用語' },
        { label: 'ペルソナ生成' },
        { label: '要約' },
        { label: '翻訳' },
        { label: 'エクスポート' }
    ];

    return (
        <div className="flex flex-col w-full p-4 gap-4 h-full">
            {/* ヘッダー部分 */}
            <div className="flex justify-between items-center w-full bg-base-100 p-4 rounded-xl shadow-sm border border-base-200 shrink-0">
                <div className="flex items-center gap-4">
                    <span className="text-2xl font-bold">翻訳プロジェクト: Skyrim.esm</span>
                    <span className="badge badge-primary badge-lg badge-outline">進行中</span>
                </div>
                <button className="btn btn-outline btn-sm">プロジェクト設定</button>
            </div>

            {/* ステッパー (全体の進捗) */}
            <div className="bg-base-100 p-4 rounded-xl shadow-sm border border-base-200 shrink-0">
                <ul className="steps w-full">
                    {tabs.map((tab, idx) => (
                        <li key={idx} className={`step ${idx <= activeTab ? 'step-primary' : ''}`}>{tab.label}</li>
                    ))}
                </ul>
            </div>

            <div className="bg-base-100 rounded-xl shadow-sm border border-base-200 flex flex-col flex-1 overflow-hidden">
                {/* タブナビゲーション */}
                <div role="tablist" className="tabs tabs-bordered w-full pt-2 shrink-0">
                    {tabs.map((tab, idx) => (
                        <button
                            key={idx}
                            role="tab"
                            className={`tab ${activeTab === idx ? 'tab-active' : ''}`}
                            onClick={() => setActiveTab(idx)}
                        >
                            {tab.label}
                        </button>
                    ))}
                </div>

                {/* ▼▼▼ タブ・コンテンツ群 ▼▼▼ */}
                <div className="flex flex-col p-4 flex-1 overflow-hidden relative">

                    {/* 1. 用語パネル (Terminology) */}
                    <TerminologyPanel isActive={activeTab === 0} onNext={() => setActiveTab(1)} />

                    {/* 2. ペルソナ生成パネル (Persona) */}
                    <PersonaPanel isActive={activeTab === 1} onNext={() => setActiveTab(2)} />

                    {/* 3. 要約パネル (Summary) */}
                    <SummaryPanel isActive={activeTab === 2} onNext={() => setActiveTab(3)} />

                    {/* 4. 翻訳パネル (Translation) */}
                    <TranslationPanel isActive={activeTab === 3} onNext={() => setActiveTab(4)} />

                    {/* 5. エクスポートパネル (Export) */}
                    <ExportPanel isActive={activeTab === 4} />

                </div>
            </div>
        </div>
    );
};

export default TranslationFlow;
