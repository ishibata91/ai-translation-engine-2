import {LoadPanel} from '../components/translation-flow/LoadPanel';
import {ExportPanel} from '../components/translation-flow/ExportPanel';
import {PersonaPanel} from '../components/translation-flow/PersonaPanel';
import {SummaryPanel} from '../components/translation-flow/SummaryPanel';
import {TerminologyPanel} from '../components/translation-flow/TerminologyPanel';
import {TranslationPanel} from '../components/translation-flow/TranslationPanel';
import {useTranslationFlow} from '../hooks/features/translationFlow/useTranslationFlow';

/**
 * 翻訳プロジェクトの工程をタブで切り替えて表示する。
 */
export default function TranslationFlow() {
    const {state, actions} = useTranslationFlow();

    return (
        <div className="flex flex-col w-full p-4 gap-4 h-full">
            <div className="flex justify-between items-center w-full bg-base-100 p-4 rounded-xl shadow-sm border border-base-200 shrink-0">
                <div className="flex items-center gap-4">
                    <span className="text-2xl font-bold">翻訳プロジェクト Task: {state.taskId}</span>
                    <span className="badge badge-primary badge-lg badge-outline">進行中</span>
                </div>
                <button className="btn btn-outline btn-sm">プロジェクト設定</button>
            </div>

            <div className="bg-base-100 p-4 rounded-xl shadow-sm border border-base-200 shrink-0">
                <ul className="steps w-full">
                    {state.tabs.map((tab, idx) => (
                        <li key={tab.label} className={`step ${idx <= state.activeTab ? 'step-primary' : ''}`}>
                            {tab.label}
                        </li>
                    ))}
                </ul>
            </div>

            <div className="bg-base-100 rounded-xl shadow-sm border border-base-200 flex flex-col flex-1 overflow-hidden">
                <div role="tablist" className="tabs tabs-bordered w-full pt-2 shrink-0">
                    {state.tabs.map((tab, idx) => (
                        <button
                            key={tab.label}
                            type="button"
                            role="tab"
                            className={`tab ${state.activeTab === idx ? 'tab-active' : ''}`}
                            onClick={() => actions.handleTabChange(idx)}
                        >
                            {tab.label}
                        </button>
                    ))}
                </div>

                <div className="flex flex-col p-4 flex-1 overflow-y-auto relative">
                    <LoadPanel
                        isActive={state.activeTab === 0}
                        selectedFiles={state.selectedFiles}
                        loadedFiles={state.loadedFiles}
                        isLoading={state.isLoading}
                        errorMessage={state.errorMessage}
                        onSelectFiles={actions.handleSelectFiles}
                        onRemoveFile={actions.handleRemoveFile}
                        onLoadSelectedFiles={actions.handleLoadSelectedFiles}
                        onReloadFiles={actions.handleReloadFiles}
                        onPreviewPageChange={actions.handlePreviewPageChange}
                        onNext={actions.handleAdvanceFromLoad}
                    />
                    <TerminologyPanel
                        isActive={state.activeTab === 1}
                        taskId={state.taskId}
                        summary={state.terminologySummary}
                        statusLabel={state.terminologyStatusLabel}
                        errorMessage={state.terminologyErrorMessage}
                        isRunning={state.isTerminologyRunning}
                        llmConfig={state.terminologyConfig}
                        promptConfig={state.terminologyPromptConfig}
                        isConfigHydrated={state.isTerminologyConfigHydrated}
                        isPromptHydrated={state.isTerminologyPromptHydrated}
                        onConfigChange={actions.handleTerminologyConfigChange}
                        onPromptChange={actions.handleTerminologyPromptChange}
                        onRun={actions.handleRunTerminologyPhase}
                        onRefresh={actions.handleRefreshTerminologyPhase}
                        onNext={actions.handleAdvanceFromTerminology}
                    />
                    <PersonaPanel isActive={state.activeTab === 2} onNext={() => actions.handleTabChange(3)} />
                    <SummaryPanel isActive={state.activeTab === 3} onNext={() => actions.handleTabChange(4)} />
                    <TranslationPanel isActive={state.activeTab === 4} onNext={() => actions.handleTabChange(5)} />
                    <ExportPanel isActive={state.activeTab === 5} />
                </div>
            </div>
        </div>
    );
}
