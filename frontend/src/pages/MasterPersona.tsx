import React, { useEffect, useState } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import ModelSettings from '../components/ModelSettings';
import DataTable from '../components/DataTable';
import PersonaDetail from '../components/PersonaDetail';
import { type NpcRow, type NpcStatus, STATUS_BADGE } from '../types/npc';
import { SelectJSONFile } from '../wailsjs/go/main/App';
import { StartMasterPersonTask } from '../wailsjs/go/task/Bridge';
import type { PhaseCompletedEvent, FrontendTask } from '../types/task';
import * as Events from '../wailsjs/runtime/runtime';

type PersonaProgressStatus = 'IN_PROGRESS' | 'COMPLETED' | 'FAILED';

interface PersonaProgressEvent {
    CorrelationID: string;
    Total: number;
    Completed: number;
    Failed: number;
    Status: PersonaProgressStatus;
    Message: string;
}

interface PersonaRequestSummary {
    request_count: number;
    npc_count: number;
}

// ── 列定義 ───────────────────────────────────────────────
const NPC_COLUMNS: ColumnDef<NpcRow, unknown>[] = [
    {
        accessorKey: 'formId',
        header: 'FormID',
        cell: (info) => <span className="font-mono text-sm">{info.getValue() as string}</span>,
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
            return <div className={`badge badge-sm ${STATUS_BADGE[s]}`}>{s}</div>;
        },
    },
    {
        accessorKey: 'updatedAt',
        header: '生成日時',
    },
];

// ── ページコンポーネント ──────────────────────────────────
const MasterPersona: React.FC = () => {
    const [npcData] = useState<NpcRow[]>([]);
    const [selectedRow, setSelectedRow] = useState<NpcRow | null>(null);
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [isGenerating, setIsGenerating] = useState<boolean>(false);
    const [jsonPath, setJsonPath] = useState<string>('');
    const [activeTaskId, setActiveTaskId] = useState<string | null>(null);
    const [progressPercent, setProgressPercent] = useState<number>(0);
    const [statusMessage, setStatusMessage] = useState<string>('待機中');
    const [errorMessage, setErrorMessage] = useState<string>('');
    const [summary, setSummary] = useState<PersonaRequestSummary | null>(null);

    const handleRowSelect = (row: NpcRow | null, rowId: string | null) => {
        setSelectedRow(row);
        setSelectedRowId(rowId);
    };

    const handlePickJson = async () => {
        const selected = await SelectJSONFile();
        if (!selected) {
            return;
        }
        setJsonPath(selected);
        setErrorMessage('');
    };

    const handleStart = async () => {
        if (!jsonPath || isGenerating) {
            return;
        }
        setIsGenerating(true);
        setActiveTaskId(null);
        setErrorMessage('');
        setSummary(null);
        setProgressPercent(0);
        setStatusMessage('タスクを開始しています...');

        try {
            const taskID = await StartMasterPersonTask({ source_json_path: jsonPath });
            setActiveTaskId(taskID);
        } catch (error) {
            setIsGenerating(false);
            setStatusMessage('タスク開始に失敗しました');
            setErrorMessage(error instanceof Error ? error.message : '不明なエラーが発生しました');
        }
    };

    useEffect(() => {
        const offProgress = Events.EventsOn('persona:progress', (event: PersonaProgressEvent) => {
            const currentTaskId = activeTaskId ?? (isGenerating ? event.CorrelationID : null);

            if (!activeTaskId && currentTaskId) {
                setActiveTaskId(currentTaskId);
            }

            if (currentTaskId && event.CorrelationID !== currentTaskId) {
                return;
            }

            if (event.Total > 0) {
                setProgressPercent(Math.min(100, Math.max(0, (event.Completed / event.Total) * 100)));
            }
            setStatusMessage(event.Message || '処理中');

            if (event.Status === 'FAILED') {
                setIsGenerating(false);
                setErrorMessage(event.Message || '生成に失敗しました');
            }
            if (event.Status === 'COMPLETED') {
                setIsGenerating(false);
            }
        });

        const offTaskUpdated = Events.EventsOn('task:updated', (task: FrontendTask) => {
            const currentTaskId = activeTaskId ?? (isGenerating ? task.id : null);

            if (!activeTaskId && currentTaskId) {
                setActiveTaskId(currentTaskId);
            }

            if (!currentTaskId || task.id !== currentTaskId) {
                return;
            }
            if (task.status === 'failed') {
                setIsGenerating(false);
                setErrorMessage(task.error_msg || 'タスク実行に失敗しました');
            }
            if (task.status === 'request_generated') {
                setIsGenerating(false);
            }
        });

        const offPhaseCompleted = Events.EventsOn('task:phase_completed', (payload: PhaseCompletedEvent) => {
            const currentTaskId = activeTaskId ?? (isGenerating ? payload.taskId : null);

            if (!activeTaskId && currentTaskId) {
                setActiveTaskId(currentTaskId);
            }

            if (!currentTaskId || payload.taskId !== currentTaskId || payload.phase !== 'REQUEST_GENERATED') {
                return;
            }
            const nextSummary = payload.summary as PersonaRequestSummary;
            setSummary({
                request_count: nextSummary.request_count ?? 0,
                npc_count: nextSummary.npc_count ?? 0,
            });
            setStatusMessage('REQUEST_GENERATED');
            setProgressPercent(100);
            setIsGenerating(false);
        });

        return () => {
            offProgress();
            offTaskUpdated();
            offPhaseCompleted();
        };
    }, [activeTaskId, isGenerating]);

    return (
        <div className="flex flex-col w-full h-full p-4 gap-4">
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
            <div className="grid grid-cols-2 gap-4 shrink-0">
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
                            <div>
                                <span className="mt-2 mb-1 block text-sm text-base-content/70 font-bold">全体進捗</span>
                                <progress className="progress progress-primary w-full" value={progressPercent} max="100"></progress>
                                <span className="text-xs text-base-content/70 mt-1 block">{statusMessage}</span>
                                {errorMessage && <span className="text-xs text-error mt-1 block">{errorMessage}</span>}
                            </div>
                        </div>
                    </div>
                </div>

                {/* 統計カード */}
                <div className="card bg-base-100 border border-base-200 shadow-sm">
                    <div className="card-body">
                        <h2 className="card-title text-base">ペルソナDB ステータス</h2>
                        <div className="grid grid-cols-2 gap-4 mt-2">
                            <div className="stat p-0">
                                <div className="stat-title text-sm">登録済みNPC数</div>
                                <div className="stat-value text-primary font-mono text-3xl">{summary?.npc_count ?? 0}</div>
                            </div>
                            <div className="stat p-0">
                                <div className="stat-title text-sm">生成リクエスト数</div>
                                <div className="stat-value text-secondary font-mono text-3xl">{summary?.request_count ?? 0}</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* モデル設定 */}
            <div className="shrink-0">
                <ModelSettings title="ペルソナ生成モデル設定" />
            </div>

            {/* 2ペインレイアウト (左: NPC テーブル, 右: PersonaDetail) */}
            <div className="flex gap-4 flex-1 min-h-0 overflow-hidden relative">
                <div className="w-1/2 flex flex-col min-h-0 border border-base-200 rounded-xl bg-base-100 overflow-hidden">
                    <DataTable
                        columns={NPC_COLUMNS}
                        data={npcData}
                        title="NPC処理ステータス (Skyrim.esm)"
                        selectedRowId={selectedRowId}
                        onRowSelect={handleRowSelect}
                        enableColumnFilter
                    />
                </div>

                <div className="w-1/2 flex flex-col min-h-0">
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
                        disabled={isGenerating || !jsonPath}
                    >
                        {isGenerating ? '生成中...' : '開始'}
                    </button>
                    <button className="btn btn-primary btn-sm" disabled={isGenerating}>
                        生成データを確定
                    </button>
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
};

export default MasterPersona;
