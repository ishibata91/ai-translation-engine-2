import {useState} from 'react';
import {ExportPanel} from '../components/translation-flow/ExportPanel';
import {PersonaPanel} from '../components/translation-flow/PersonaPanel';
import {SummaryPanel} from '../components/translation-flow/SummaryPanel';
import {TerminologyPanel} from '../components/translation-flow/TerminologyPanel';
import {TranslationPanel} from '../components/translation-flow/TranslationPanel';

const TABS = [
    { label: '用語' },
    { label: 'ペルソナ生成' },
    { label: '要約' },
    { label: '翻訳' },
    { label: 'エクスポート' },
];

/**
 * 翻訳プロジェクトの工程をタブで切り替えて表示する。
 */
export default function TranslationFlow() {
    const [activeTab, setActiveTab] = useState(0);

    return (
        <div className="flex flex-col w-full p-4 gap-4 h-full">
            <div className="flex justify-between items-center w-full bg-base-100 p-4 rounded-xl shadow-sm border border-base-200 shrink-0">
                <div className="flex items-center gap-4">
                    <span className="text-2xl font-bold">翻訳プロジェクト: Skyrim.esm</span>
                    <span className="badge badge-primary badge-lg badge-outline">進行中</span>
                </div>
                <button className="btn btn-outline btn-sm">プロジェクト設定</button>
            </div>

            <div className="bg-base-100 p-4 rounded-xl shadow-sm border border-base-200 shrink-0">
                <ul className="steps w-full">
                    {TABS.map((tab, idx) => (
                        <li key={tab.label} className={`step ${idx <= activeTab ? 'step-primary' : ''}`}>
                            {tab.label}
                        </li>
                    ))}
                </ul>
            </div>

            <div className="bg-base-100 rounded-xl shadow-sm border border-base-200 flex flex-col flex-1 overflow-hidden">
                <div role="tablist" className="tabs tabs-bordered w-full pt-2 shrink-0">
                    {TABS.map((tab, idx) => (
                        <button
                            key={tab.label}
                            type="button"
                            role="tab"
                            className={`tab ${activeTab === idx ? 'tab-active' : ''}`}
                            onClick={() => setActiveTab(idx)}
                        >
                            {tab.label}
                        </button>
                    ))}
                </div>

                <div className="flex flex-col p-4 flex-1 overflow-hidden relative">
                    <TerminologyPanel isActive={activeTab === 0} onNext={() => setActiveTab(1)} />
                    <PersonaPanel isActive={activeTab === 1} onNext={() => setActiveTab(2)} />
                    <SummaryPanel isActive={activeTab === 2} onNext={() => setActiveTab(3)} />
                    <TranslationPanel isActive={activeTab === 3} onNext={() => setActiveTab(4)} />
                    <ExportPanel isActive={activeTab === 4} />
                </div>
            </div>
        </div>
    );
}
