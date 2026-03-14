import React, {useState} from 'react';
import type {NpcRow} from '../types/npc';

export type DetailTab = 'persona' | 'dialogues' | 'request' | 'meta';

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
        <div className="flex flex-col gap-4 h-full w-full bg-base-100 border border-base-200 rounded-xl p-4 overflow-y-auto">
            {/* ヘッダー */}
            <div className="shrink-0 flex justify-between items-start border-b pb-2">
                <div>
                    <h3 className="text-xl font-bold">{npc.name} <span className="text-base text-base-content/60 font-mono">({npc.formId})</span></h3>
                </div>
            </div>

            {/* タブナビゲーション */}
            <div className="tabs tabs-boxed bg-base-200 w-full shrink-0">
                <a className={`tab flex-1 ${detailTab === 'persona' ? 'tab-active' : ''}`} onClick={() => setDetailTab('persona')}>ペルソナ</a>
                <a className={`tab flex-1 ${detailTab === 'dialogues' ? 'tab-active' : ''}`} onClick={() => setDetailTab('dialogues')}>セリフ一覧</a>
                <a className={`tab flex-1 ${detailTab === 'request' ? 'tab-active' : ''}`} onClick={() => setDetailTab('request')}>生成リクエスト</a>
                <a className={`tab flex-1 ${detailTab === 'meta' ? 'tab-active' : ''}`} onClick={() => setDetailTab('meta')}>メタ情報</a>
            </div>

            {/* ペルソナタブ */}
            {detailTab === 'persona' && (
                <div className="form-control flex-1 flex flex-col min-h-0 gap-2 w-full">
                    <label className="label"><span className="label-text font-bold">生成されたペルソナ情報 (プロンプト動的注入用)</span></label>
                    <textarea
                        className="textarea textarea-bordered flex-1 text-base leading-relaxed w-full"
                        value={npc.personaText}
                        readOnly
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
                                    <th className="w-20 sm:w-28 md:w-40">EditorID</th>
                                    <th className="w-full">原文</th>
                                </tr>
                            </thead>
                            <tbody>
                                {npc.dialogues.map((d, i) => (
                                    <tr key={i}>
                                        <td
                                            className="font-mono text-xs max-w-[5rem] sm:max-w-[7rem] md:max-w-[10rem] truncate"
                                            title={d.editorId}
                                        >
                                            {d.editorId}
                                        </td>
                                        <td className="text-base-content/70 whitespace-normal min-w-[200px] break-words">
                                            {d.source}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    ) : (
                        <div className="text-sm text-base-content/40 p-4">(セリフデータなし)</div>
                    )}
                </div>
            )}

            {/* 生成リクエストタブ */}
            {detailTab === 'request' && (
                <div className="mockup-code flex-1 overflow-y-auto bg-base-200 text-base-content text-sm border border-base-300">
                    <pre data-prefix=">"><code>{npc.generationRequest || '(生成リクエストなし)'}</code></pre>
                </div>
            )}

            {/* メタ情報タブ */}
            {detailTab === 'meta' && (
                <div className="flex flex-col gap-4 text-sm flex-1 overflow-y-auto min-h-0 p-4 border rounded-lg bg-base-50">
                    <div className="grid grid-cols-[120px_1fr] gap-4 items-center">
                        <span className="font-bold">FormID</span>
                        <span className="font-mono">{npc.formId}</span>

                        <span className="font-bold">PersonaID</span>
                        <span className="font-mono">{npc.personaId}</span>

                        <span className="font-bold">NPC名</span>
                        <span>{npc.name}</span>

                        <span className="font-bold">Race</span>
                        <span>{npc.race || 'Unknown'}</span>

                        <span className="font-bold">Sex</span>
                        <span>{npc.sex || 'Unknown'}</span>

                        <span className="font-bold">Voice</span>
                        <span>{npc.voiceType || 'Unknown'}</span>

                        <span className="font-bold">Source Plugin</span>
                        <span className="font-mono">{npc.sourcePlugin || 'UNKNOWN'}</span>

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
