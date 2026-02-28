import React, { useState } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import ModelSettings from '../components/ModelSettings';
import DataTable from '../components/DataTable';
import DetailPane from '../components/DetailPane';

// ── 型定義 ──────────────────────────────────────────────
type NpcStatus = '完了' | '生成中' | '抽出完了' | 'エラー';

interface Dialogue {
    recordType: string;
    editorId: string;
    source: string;
    translation: string;
}

interface NpcRow {
    formId: string;
    name: string;
    dialogueCount: number;
    status: NpcStatus;
    updatedAt: string;
    promptHistory: string[];
    rawResponse: string;
    dialogues: Dialogue[];
}

// ── モックデータ ─────────────────────────────────────────
const NPC_DATA: NpcRow[] = [
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

// ── ステータスバッジ ──────────────────────────────────────
const STATUS_BADGE: Record<NpcStatus, string> = {
    '完了': 'badge-success',
    '生成中': 'badge-info',
    '抽出完了': 'badge-ghost',
    'エラー': 'badge-error',
};

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

type DetailTab = 'dialogues' | 'prompt' | 'raw' | 'meta';

// ── ページコンポーネント ──────────────────────────────────
const MasterPersona: React.FC = () => {
    const [selectedRow, setSelectedRow] = useState<NpcRow | null>(null);
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [detailTab, setDetailTab] = useState<DetailTab>('dialogues');

    const handleRowSelect = (row: NpcRow | null, rowId: string | null) => {
        setSelectedRow(row);
        setSelectedRowId(rowId);
        if (row) setDetailTab('dialogues');
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

            {/* NPC テーブル (flex-1 で残り縦幅を占有) */}
            <DataTable
                columns={NPC_COLUMNS}
                data={NPC_DATA}
                title="NPC処理ステータス (Skyrim.esm)"
                selectedRowId={selectedRowId}
                onRowSelect={handleRowSelect}
                enableColumnFilter
            />

            {/* 詳細ペイン (選択行があると下からスライドイン) */}
            <DetailPane
                isOpen={!!selectedRow}
                onClose={() => handleRowSelect(null, null)}
                title={selectedRow ? `詳細: ${selectedRow.name} (${selectedRow.formId})` : '詳細'}
                defaultHeight={280}
            >
                {selectedRow && (
                    <div className="flex flex-col gap-3 h-full">
                        {/* タブ */}
                        <div className="tabs tabs-boxed bg-base-200 w-fit shrink-0">
                            <a className={`tab ${detailTab === 'dialogues' ? 'tab-active' : ''}`} onClick={() => setDetailTab('dialogues')}>セリフ一覧</a>
                            <a className={`tab ${detailTab === 'prompt' ? 'tab-active' : ''}`} onClick={() => setDetailTab('prompt')}>プロンプト履歴</a>
                            <a className={`tab ${detailTab === 'raw' ? 'tab-active' : ''}`} onClick={() => setDetailTab('raw')}>RAWレスポンス</a>
                            <a className={`tab ${detailTab === 'meta' ? 'tab-active' : ''}`} onClick={() => setDetailTab('meta')}>メタ情報</a>
                        </div>

                        {/* セリフ一覧タブ */}
                        {detailTab === 'dialogues' && (
                            <div className="flex-1 overflow-y-auto min-h-0">
                                {selectedRow.dialogues.length > 0 ? (
                                    <table className="table table-zebra table-pin-rows w-full text-sm">
                                        <thead>
                                            <tr>
                                                <th className="w-16">Type</th>
                                                <th className="w-48">EditorID</th>
                                                <th>原文</th>
                                                <th>訳文</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {selectedRow.dialogues.map((d, i) => (
                                                <tr key={i}>
                                                    <td><div className="badge badge-outline badge-sm font-mono">{d.recordType}</div></td>
                                                    <td className="font-mono text-xs">{d.editorId}</td>
                                                    <td className="text-base-content/70">{d.source}</td>
                                                    <td>{d.translation}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                ) : (
                                    <div className="text-sm text-base-content/40 p-2">(セリフデータなし)</div>
                                )}
                            </div>
                        )}

                        {/* プロンプト履歴タブ */}
                        {detailTab === 'prompt' && (
                            <div className="mockup-code flex-1 overflow-y-auto bg-base-200 text-base-content text-sm">
                                {selectedRow.promptHistory.length > 0
                                    ? selectedRow.promptHistory.map((line, i) => (
                                        <pre key={i} data-prefix={i + 1}><code>{line}</code></pre>
                                    ))
                                    : <pre data-prefix="—"><code className="text-base-content/40">(プロンプト履歴なし)</code></pre>
                                }
                            </div>
                        )}

                        {/* RAWレスポンスタブ */}
                        {detailTab === 'raw' && (
                            <div className="mockup-code flex-1 overflow-y-auto bg-base-200 text-base-content text-sm">
                                <pre data-prefix=">"><code>{selectedRow.rawResponse}</code></pre>
                            </div>
                        )}

                        {/* メタ情報タブ */}
                        {detailTab === 'meta' && (
                            <div className="flex flex-col gap-2 text-sm">
                                <div className="flex gap-4">
                                    <span className="font-bold w-32">FormID</span>
                                    <span className="font-mono">{selectedRow.formId}</span>
                                </div>
                                <div className="flex gap-4">
                                    <span className="font-bold w-32">NPC名</span>
                                    <span>{selectedRow.name}</span>
                                </div>
                                <div className="flex gap-4">
                                    <span className="font-bold w-32">ステータス</span>
                                    <div className={`badge badge-sm ${STATUS_BADGE[selectedRow.status]}`}>{selectedRow.status}</div>
                                </div>
                                <div className="flex gap-4">
                                    <span className="font-bold w-32">セリフ数</span>
                                    <span className="font-mono">{selectedRow.dialogueCount} 行</span>
                                </div>
                                <div className="flex gap-4">
                                    <span className="font-bold w-32">生成日時</span>
                                    <span>{selectedRow.updatedAt}</span>
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </DetailPane>

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
