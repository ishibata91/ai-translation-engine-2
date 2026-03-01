import React, { useState } from 'react';
import type { NpcRow } from '../types/npc';
import { STATUS_BADGE } from '../types/npc';

export type DetailTab = 'persona' | 'dialogues' | 'prompt' | 'raw' | 'meta';

interface PersonaDetailProps {
    npc: NpcRow | null;
}

const PersonaDetail: React.FC<PersonaDetailProps> = ({ npc }) => {
    const [detailTab, setDetailTab] = useState<DetailTab>('persona');

    if (!npc) {
        return (
            <div className="flex items-center justify-center h-full text-base-content/40 bg-base-100 rounded-xl border border-base-200">
                表示するNPCを選択してください
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-4 h-full bg-base-100 border border-base-200 rounded-xl p-4 overflow-y-auto">
            {/* ヘッダー */}
            <div className="shrink-0 flex justify-between items-start border-b pb-2">
                <div>
                    <h3 className="text-xl font-bold">{npc.name} <span className="text-base text-base-content/60 font-mono">({npc.formId})</span></h3>
                    <div className="flex gap-2 mt-2">
                        <span className="badge badge-outline">Race: Nord</span>
                        <span className="badge badge-outline">Sex: Male</span>
                        <span className="badge badge-outline">Class: Warrior</span>
                    </div>
                </div>
                <div className={`badge badge-sm ${STATUS_BADGE[npc.status]}`}>{npc.status}</div>
            </div>

            {/* タブナビゲーション */}
            <div className="tabs tabs-boxed bg-base-200 w-fit shrink-0">
                <a className={`tab ${detailTab === 'persona' ? 'tab-active' : ''}`} onClick={() => setDetailTab('persona')}>ペルソナ</a>
                <a className={`tab ${detailTab === 'dialogues' ? 'tab-active' : ''}`} onClick={() => setDetailTab('dialogues')}>セリフ一覧</a>
                <a className={`tab ${detailTab === 'prompt' ? 'tab-active' : ''}`} onClick={() => setDetailTab('prompt')}>プロンプト履歴</a>
                <a className={`tab ${detailTab === 'raw' ? 'tab-active' : ''}`} onClick={() => setDetailTab('raw')}>RAWレスポンス</a>
                <a className={`tab ${detailTab === 'meta' ? 'tab-active' : ''}`} onClick={() => setDetailTab('meta')}>メタ情報</a>
            </div>

            {/* ペルソナタブ */}
            {detailTab === 'persona' && (
                <div className="form-control flex-1 flex flex-col min-h-0 gap-2">
                    <label className="label"><span className="label-text font-bold">生成されたペルソナ情報 (プロンプト動的注入用)</span></label>
                    <textarea
                        className="textarea textarea-bordered flex-1 text-base leading-relaxed"
                        defaultValue={`誇り高く、カリスマ性のあるストームクロークの反乱軍リーダー。雄弁で威厳のある口調。\n一人称: 私、俺\n二人称: お前、貴様\n特徴: スカイリムの独立とノルドの誇りを強調する。「～だ」「～だろう」「～なのだ」といった断定的で力強い語尾を多用する。若者には厳しくも期待を込めて接する。`}
                    />
                    <div className="flex justify-end gap-2 mt-2 shrink-0">
                        <button className="btn btn-outline btn-sm">再生成</button>
                        <button className="btn btn-secondary btn-sm">保存</button>
                    </div>
                </div>
            )}

            {/* セリフ一覧タブ */}
            {detailTab === 'dialogues' && (
                <div className="flex-1 overflow-y-auto min-h-0 border rounded-lg border-base-200">
                    {npc.dialogues.length > 0 ? (
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
                                {npc.dialogues.map((d, i) => (
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
                        <div className="text-sm text-base-content/40 p-4">(セリフデータなし)</div>
                    )}
                </div>
            )}

            {/* プロンプト履歴タブ */}
            {detailTab === 'prompt' && (
                <div className="mockup-code flex-1 overflow-y-auto bg-base-200 text-base-content text-sm border border-base-300">
                    {npc.promptHistory.length > 0
                        ? npc.promptHistory.map((line, i) => (
                            <pre key={i} data-prefix={i + 1}><code>{line}</code></pre>
                        ))
                        : <pre data-prefix="—"><code className="text-base-content/40">(プロンプト履歴なし)</code></pre>
                    }
                </div>
            )}

            {/* RAWレスポンスタブ */}
            {detailTab === 'raw' && (
                <div className="mockup-code flex-1 overflow-y-auto bg-base-200 text-base-content text-sm border border-base-300">
                    <pre data-prefix=">"><code>{npc.rawResponse}</code></pre>
                </div>
            )}

            {/* メタ情報タブ */}
            {detailTab === 'meta' && (
                <div className="flex flex-col gap-4 text-sm flex-1 overflow-y-auto min-h-0 p-4 border rounded-lg bg-base-50">
                    <div className="grid grid-cols-[120px_1fr] gap-4 items-center">
                        <span className="font-bold">FormID</span>
                        <span className="font-mono">{npc.formId}</span>

                        <span className="font-bold">NPC名</span>
                        <span>{npc.name}</span>

                        <span className="font-bold">ステータス</span>
                        <div><div className={`badge badge-sm ${STATUS_BADGE[npc.status]}`}>{npc.status}</div></div>

                        <span className="font-bold">セリフ数</span>
                        <span className="font-mono">{npc.dialogueCount} 行</span>

                        <span className="font-bold">生成日時</span>
                        <span>{npc.updatedAt}</span>
                    </div>
                </div>
            )}
        </div>
    );
};

export default PersonaDetail;
