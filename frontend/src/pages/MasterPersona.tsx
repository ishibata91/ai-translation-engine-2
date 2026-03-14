import type {ColumnDef} from '@tanstack/react-table';
import ModelSettings from '../components/ModelSettings';
import DataTable from '../components/DataTable';
import PersonaDetail from '../components/PersonaDetail';
import PromptSettingCard from '../components/masterPersona/PromptSettingCard';
import {NPC_STATUS_LABEL, type NpcRow, type NpcStatus, STATUS_BADGE} from '../types/npc';
import {useMasterPersona} from '../hooks/features/masterPersona/useMasterPersona';


// ── 列定義 ───────────────────────────────────────────────
const NPC_COLUMNS: ColumnDef<NpcRow, unknown>[] = [
    {
        accessorKey: 'formId',
        header: 'FormID',
        cell: (info) => <span className="font-mono text-sm">{info.getValue() as string}</span>,
    },
    {
        accessorKey: 'sourcePlugin',
        header: 'プラグイン名',
        cell: (info) => <span className="font-mono text-xs">{info.getValue() as string}</span>,
    },
    {
        accessorKey: 'name',
        header: 'NPC名 (EditorID)',
    },
    {
        accessorKey: 'dialogueCount',
        header: 'セリフ数',
        cell: (info) => <span className="font-mono text-right block">{info.getValue() as number}</span>,
    },
    {
        accessorKey: 'status',
        header: 'ステータス',
        cell: (info) => {
            const s = info.getValue() as NpcStatus;
            return <div className={`badge badge-sm ${STATUS_BADGE[s]}`}>{NPC_STATUS_LABEL[s]}</div>;
        },
    },
    {
        accessorKey: 'updatedAt',
        header: '生成日時',
    },
];

// ── ページコンポーネント ──────────────────────────────────
/**
 * マスターペルソナの生成設定と結果確認画面を描画する。
 */
export default function MasterPersona() {
    const {
        selectedRow,
        selectedRowId,
        npcSearchInput,
        setNpcSearchInput,
        pluginFilterInput,
        setPluginFilterInput,
        statusFilterInput,
        setStatusFilterInput,
        npcPage,
        setNpcPage,
        isGenerating,
        jsonPath,
        overwriteExisting,
        setOverwriteExisting,
        activeTaskId,
        progressPercent,
        progressMode,
        progressPrimaryMessage,
        progressSecondaryMessage,
        isProgressIndeterminate,
        errorMessage,
        progressCounts,
        activeTaskStatus,
        llmConfig,
        setLLMConfig,
        isLLMConfigHydrated,
        isModelSettingsLocked,
        promptConfig,
        setPromptConfig,
        isPromptConfigHydrated,
        pluginOptions,
        filteredNpcData,
        pagedNpcData,
        totalNpcPages,
        handleRowSelect,
        applyNPCFilters,
        clearNPCFilters,
        handlePickJson,
        handleStart,
        handleResumeCurrentTask,
        handlePauseCurrentTask,
        providerNamespace,
    } = useMasterPersona();

    return (
        <div className="flex flex-col w-full min-h-full p-4 gap-4">
            {/* ヘッダー */}
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4 shrink-0">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">マスターペルソナ構築 (Master Persona Builder)</span>
                </div>
            </div>

            {/* 通知エリア */}
            <div className="alert alert-info shadow-sm shrink-0">
                <span><code>extractData.pas</code> で抽出されたベースゲームのJSONデータからNPCのセリフを解析し、LLMを用いて基本となるペルソナ（性格・口調）を生成・キャッシュします。これによりMod翻訳時の品質と一貫性が向上します。</span>
            </div>

            {/* 上部パネル */}
            <div className="grid grid-cols-1 gap-4 shrink-0">
                {/* 生成設定カード */}
                <div className="card bg-base-100 border border-base-200 shadow-sm">
                    <div className="card-body">
                        <h2 className="card-title text-base">JSONデータのインポートと生成</h2>
                        <div className="flex flex-col gap-4 mt-2">
                            <span className="text-sm">xEditスクリプト <code>extractData.pas</code> によって抽出された、マスターファイルのJSONデータを選択し、ペルソナ生成を開始します。</span>
                            <div className="flex gap-4 items-center">
                                <input type="text" readOnly value={jsonPath} placeholder="JSONファイルを選択してください" className="input input-bordered w-full max-w-xl font-mono text-xs" />
                                <button className="btn btn-outline btn-primary" onClick={handlePickJson}>JSON選択</button>
                            </div>
                            <label className="label cursor-pointer justify-start gap-3">
                                <input
                                    type="checkbox"
                                    className="checkbox checkbox-primary checkbox-sm"
                                    checked={overwriteExisting}
                                    onChange={(event) => setOverwriteExisting(event.target.checked)}
                                    disabled={isGenerating}
                                />
                                <span className="label-text">重複時に既存ペルソナを上書きする</span>
                            </label>
                            <div>
                                <span className="mt-2 mb-1 block text-sm text-base-content/70 font-bold">
                                    {progressMode === 'remote' ? 'クラウド実行進捗' : '全体進捗'}
                                </span>
                                {isProgressIndeterminate ? (
                                    <progress className="progress progress-primary w-full"></progress>
                                ) : (
                                    <progress className="progress progress-primary w-full" value={progressPercent} max="100"></progress>
                                )}
                                {progressCounts.total > 0 && (
                                    <span className="text-xs text-base-content/70 mt-1 block">
                                        {progressCounts.current} / {progressCounts.total} 件
                                    </span>
                                )}
                                <span className="text-xs text-base-content/70 mt-1 block">{progressPrimaryMessage}</span>
                                {progressSecondaryMessage !== '' && (
                                    <span className="text-xs text-base-content/60 mt-1 block">{progressSecondaryMessage}</span>
                                )}
                                {isProgressIndeterminate && (
                                    <span className="text-xs text-base-content/60 mt-1 block">
                                        進捗が取得できないため、不定表示で更新中です。
                                    </span>
                                )}
                                {errorMessage && <span className="text-xs text-error mt-1 block">{errorMessage}</span>}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Prompt 設定 */}
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 shrink-0">
                {isPromptConfigHydrated ? (
                    <>
                        <PromptSettingCard
                            title="ユーザープロンプト"
                            description="可変の指示だけを編集します。NPC メタデータや会話履歴は送信時に別途付与されます。"
                            value={promptConfig.userPrompt}
                            onChange={(value) => setPromptConfig((prev) => ({ ...prev, userPrompt: value }))}
                            badgeLabel="編集可能"
                        />
                        <PromptSettingCard
                            title="システムプロンプト"
                            description="固定の分析ルールと出力形式です。画面表示と実際の送信内容は同じ system prompt を参照します。"
                            value={promptConfig.systemPrompt}
                            readOnly
                            badgeLabel="Read Only"
                        />
                    </>
                ) : (
                    <div className="card bg-base-100 border border-base-200 shadow-sm xl:col-span-2">
                        <div className="card-body py-4">
                            <span className="text-sm text-base-content/60">プロンプト設定を読み込み中...</span>
                        </div>
                    </div>
                )}
            </div>

            {/* モデル設定 */}
            <div className="shrink-0">
                {isLLMConfigHydrated ? (
                    <ModelSettings
                        title="ペルソナ生成モデル設定"
                        value={llmConfig}
                        onChange={setLLMConfig}
                        enabled={isLLMConfigHydrated}
                        namespace={providerNamespace(llmConfig.provider)}
                        locked={isModelSettingsLocked}
                    />
                ) : (
                    <div className="card bg-base-100 border border-base-200 shadow-sm">
                        <div className="card-body py-4">
                            <span className="text-sm text-base-content/60">モデル設定を読み込み中...</span>
                        </div>
                    </div>
                )}
            </div>

            {/* 2ペインレイアウト (左: NPC テーブル, 右: PersonaDetail) */}
            <div className="flex gap-4 flex-1 min-h-[500px] overflow-hidden relative">
                <div className="w-1/2 flex flex-col min-h-[500px] border border-base-200 rounded-xl bg-base-100 overflow-hidden">
                    <DataTable
                        columns={NPC_COLUMNS}
                        data={pagedNpcData}
                        title="ペルソナ一覧"
                        headerActions={
                            <div className="flex flex-wrap items-center gap-2">
                                <input
                                    type="text"
                                    className="input input-bordered input-xs w-44"
                                    placeholder="NPC / FormID / ペルソナ検索"
                                    value={npcSearchInput}
                                    onChange={(event) => setNpcSearchInput(event.target.value)}
                                />
                                <select
                                    className="select select-bordered select-xs w-40"
                                    value={pluginFilterInput}
                                    onChange={(event) => setPluginFilterInput(event.target.value)}
                                >
                                    <option value="">全プラグイン</option>
                                    {pluginOptions.map((plugin) => (
                                        <option key={plugin} value={plugin}>{plugin}</option>
                                    ))}
                                </select>
                                <select
                                    className="select select-bordered select-xs w-32"
                                    value={statusFilterInput}
                                    onChange={(event) => setStatusFilterInput(event.target.value)}
                                >
                                    <option value="">全状態</option>
                                    <option value="draft">{NPC_STATUS_LABEL.draft}</option>
                                    <option value="generated">{NPC_STATUS_LABEL.generated}</option>
                                </select>
                                <button className="btn btn-primary btn-xs" onClick={applyNPCFilters}>
                                    検索
                                </button>
                                <button className="btn btn-ghost btn-xs" onClick={clearNPCFilters}>
                                    解除
                                </button>
                                <span className="text-xs text-base-content/60">
                                    {filteredNpcData.length.toLocaleString()} 件 / {npcPage} / {totalNpcPages} ページ
                                </span>
                                <button
                                    className="btn btn-outline btn-xs"
                                    disabled={npcPage <= 1}
                                    onClick={() => setNpcPage((prev) => Math.max(1, prev - 1))}
                                >
                                    前へ
                                </button>
                                <button
                                    className="btn btn-outline btn-xs"
                                    disabled={npcPage >= totalNpcPages}
                                    onClick={() => setNpcPage((prev) => Math.min(totalNpcPages, prev + 1))}
                                >
                                    次へ
                                </button>
                            </div>
                        }
                        selectedRowId={selectedRowId}
                        onRowSelect={handleRowSelect}
                    />
                </div>

                <div className="w-1/2 flex flex-col min-h-[500px]">
                    <PersonaDetail npc={selectedRow} />
                </div>

                {isGenerating && (
                    <div className="absolute inset-0 bg-base-100/50 backdrop-blur-[1px] z-10 flex flex-col items-center justify-center gap-4 rounded-xl border border-base-200">
                        <span className="loading loading-spinner text-primary loading-lg"></span>
                        <div className="flex flex-col items-center gap-1">
                            <span className="font-bold text-lg text-base-content/70">マスターペルソナを一括生成中...</span>
                            <span className="text-sm text-base-content/50">選択されたJSONデータから全NPCのセリフを解析しています</span>
                        </div>
                    </div>
                )}
            </div>

            {/* 下部ステータスバー */}
            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0">
                <span className="text-sm font-bold text-gray-500 ml-2">Job: MasterPersonaGeneration ({isGenerating ? 'Running' : 'Stopped'})</span>
                <div className="flex gap-2">
                    <button
                        className={`btn btn-sm ${isGenerating ? 'btn-ghost' : 'btn-outline'}`}
                        onClick={handleStart}
                        disabled={isGenerating || !jsonPath || activeTaskStatus === 'running'}
                    >
                        {isGenerating ? '生成中...' : '新規タスク開始'}
                    </button>
                    {activeTaskId && activeTaskStatus !== 'running' && (
                        <button
                            className="btn btn-success btn-sm"
                            onClick={handleResumeCurrentTask}
                        >
                            生成開始
                        </button>
                    )}
                    {activeTaskId && activeTaskStatus === 'running' && (
                        <button
                            className="btn btn-warning btn-sm"
                            onClick={handlePauseCurrentTask}
                        >
                            一時停止
                        </button>
                    )}
                </div>
            </div>

            {/* ペルソナ確認モーダル */}
            <dialog id="persona_modal" className="modal">
                <div className="modal-box w-11/12 max-w-3xl border border-base-300">
                    <h3 className="font-bold text-lg">ペルソナ詳細確認: UlfricStormcloak (00013B9B)</h3>
                    <div className="py-4 flex flex-col gap-4">
                        <div className="form-control">
                            <label className="label"><span className="label-text font-bold">要約 (Summary)</span></label>
                            <textarea className="textarea textarea-bordered h-24" readOnly value="ストームクロークの反乱軍のリーダー。誇り高く、ノルドの伝統を重んじる。ウィンドヘルムの首長であり、帝国に強い敵対心を抱いている。"></textarea>
                        </div>
                        <div className="form-control">
                            <label className="label"><span className="label-text font-bold">口調・一人称・二人称 (Tone/Pronouns)</span></label>
                            <textarea className="textarea textarea-bordered h-24" readOnly value={`一人称：「俺」\n二人称：「お前」「お前たち」\n口調：威厳があり、力強く、少し乱暴な言葉遣い。命令形をよく使う。`}></textarea>
                        </div>
                    </div>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-outline">再生成</button>
                                <button className="btn btn-primary">閉じる</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>

            {/* 削除確認モーダル */}
            <dialog id="delete_modal" className="modal">
                <div className="modal-box border border-error">
                    <h3 className="font-bold text-lg text-error">削除の確認</h3>
                    <p className="py-4">このNPCデータを一覧およびデータベースから削除しますか？<br />※この操作は取り消せません。</p>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-ghost">キャンセル</button>
                                <button className="btn btn-error">削除する</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>
        </div>
    );
}
