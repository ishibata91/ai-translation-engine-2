import React, { useState } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import ModelSettings from '../components/ModelSettings';
import DataTable from '../components/DataTable';
import PersonaDetail from '../components/PersonaDetail';
import { type NpcRow, type NpcStatus, STATUS_BADGE } from '../types/npc';
// ── モックデータ ─────────────────────────────────────────
export const NPC_DATA: NpcRow[] = [
    {
        formId: '00013B9B', name: 'UlfricStormcloak', dialogueCount: 342, status: '完了',
        updatedAt: '2026-02-26 14:02',
        promptHistory: [
            '[SYSTEM] You are a localization specialist for Skyrim...',
            '[USER] Translate NPC dialogue. Speaker: UlfricStormcloak. Race: Nord. Voice: MaleEvenToned.',
            '[ASSISTANT] 「ドラゴンボーンよ、俺と共にスカイリムのために戦え。」',
        ],
        rawResponse: '{"id":"chatcmpl-abc123","model":"gemini-2.0-flash","usage":{"prompt_tokens":1240,"completion_tokens":32},"choices":[{"message":{"content":"「ドラゴンボーンよ、俺と共にスカイリムのために戦え。」"}}]}',
        dialogues: [
            { recordType: 'INFO', editorId: 'DialogueUlfric001', source: "What is it? I'm in the middle of something.", translation: '何だ？今、手が離せないんだ。' },
            { recordType: 'INFO', editorId: 'DialogueUlfric002', source: 'Victory or Sovngarde!', translation: '勝利か、ソブンガルデか！' },
            { recordType: 'INFO', editorId: 'DialogueUlfric003', source: 'The Reach was ours, and it will be ours again.', translation: 'リーチは俺たちのものだった。そしてまた俺たちのものになる。' },
            { recordType: 'DIAL', editorId: 'DialogueUlfricGreeting', source: 'Ulfric Stormcloak, at your service.', translation: 'ウルフリック・ストームクロークだ。世話になろう。' },
            { recordType: 'INFO', editorId: 'DialogueUlfric004', source: 'Talos be with you.', translation: 'タロスのお在りを。' },
        ],
    },
    {
        formId: '0001A694', name: 'Tullius', dialogueCount: 298, status: '完了',
        updatedAt: '2026-02-26 14:05',
        promptHistory: [
            '[SYSTEM] You are a localization specialist for Skyrim...',
            '[USER] Translate NPC dialogue. Speaker: Tullius. Race: Imperial. Voice: MaleCommander.',
            '[ASSISTANT] 「この反乱はすぐに終わらせる。帝国の意志は揺るがない。」',
        ],
        rawResponse: '{"id":"chatcmpl-def456","model":"gemini-2.0-flash","usage":{"prompt_tokens":1180,"completion_tokens":28},"choices":[{"message":{"content":"「この反乱はすぐに終わらせる。帝国の意志は揺るがない。」"}}]}',
        dialogues: [
            { recordType: 'INFO', editorId: 'DialogueTullius001', source: 'Rikke, get these men moving!', translation: 'リッケ、部下たちを動かせ！' },
            { recordType: 'INFO', editorId: 'DialogueTullius002', source: 'For the Empire!', translation: '帝国のために！' },
            { recordType: 'DIAL', editorId: 'DialogueTulliusGreeting', source: 'General Tullius. Commander of the Imperial Legion in Skyrim.', translation: 'タリアス将軍。スカイリム駐屯する帝国軍団の司令官だ。' },
        ],
    },
    {
        formId: '0001A695', name: 'Rikke', dialogueCount: 156, status: '生成中',
        updatedAt: '-',
        promptHistory: ['[SYSTEM] You are a localization specialist for Skyrim...'],
        rawResponse: '(生成中...)',
        dialogues: [
            { recordType: 'INFO', editorId: 'DialogueRikke001', source: 'Yes, sir!', translation: 'はい、将軍！' },
        ],
    },
    {
        formId: '00013BA1', name: 'BalgruufTheGreater', dialogueCount: 210, status: '抽出完了',
        updatedAt: '-',
        promptHistory: [],
        rawResponse: '(未実行)',
        dialogues: [
            { recordType: 'INFO', editorId: 'DialogueBalgruuf001', source: 'Ah, Dragonborn! Good to see you.', translation: 'ああ、ドラゴンボーン！会えて嬉しいぞ。' },
            { recordType: 'INFO', editorId: 'DialogueBalgruuf002', source: "Whiterun's walls will hold, I assure you.", translation: 'ホワイトランの城壁は持たせる。保証する。' },
            { recordType: 'DIAL', editorId: 'DialogueBalgruufAngle', source: 'What brings you to Dragonsreach?', translation: 'ドラゴンリーチに何の用か？' },
        ],
    },
    {
        formId: '0001A696', name: 'GalmarStoneFist', dialogueCount: 124, status: 'エラー',
        updatedAt: '2026-02-26 14:10',
        promptHistory: [
            '[SYSTEM] You are a localization specialist for Skyrim...',
            '[USER] Translate NPC dialogue. Speaker: GalmarStoneFist.',
        ],
        rawResponse: '{"error":{"code":429,"message":"Rate limit exceeded. Retry after 60 seconds."}}',
        dialogues: [
            { recordType: 'INFO', editorId: 'DialogueGalmar001', source: 'We fight for Skyrim!', translation: '向かうはスカイリムのために戦う！' },
        ],
    },
];

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
    const [selectedRow, setSelectedRow] = useState<NpcRow | null>(null);
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [isGenerating, setIsGenerating] = useState<boolean>(true);

    const handleRowSelect = (row: NpcRow | null, rowId: string | null) => {
        setSelectedRow(row);
        setSelectedRowId(rowId);
    };


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
                                <input type="file" accept=".json" className="file-input file-input-bordered file-input-primary w-full max-w-xs" />
                                <button className="btn btn-primary">アップロード</button>
                            </div>
                            <div>
                                <span className="mt-2 mb-1 block text-sm text-base-content/70 font-bold">全体進捗</span>
                                <progress className="progress progress-primary w-full" value="45" max="100"></progress>
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
                                <div className="stat-value text-primary font-mono text-3xl">2,451</div>
                            </div>
                            <div className="stat p-0">
                                <div className="stat-title text-sm">生成エラー</div>
                                <div className="stat-value text-error font-mono text-3xl">12</div>
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
                        data={NPC_DATA}
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
                        onClick={() => setIsGenerating(!isGenerating)}
                    >
                        {isGenerating ? '一時停止' : '再開 (デモ)'}
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
