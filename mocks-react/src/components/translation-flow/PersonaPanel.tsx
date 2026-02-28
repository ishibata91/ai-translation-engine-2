import React, { useState } from 'react';
import ModelSettings from '../ModelSettings';
import PersonaDetail from '../PersonaDetail';
import { NPC_DATA } from '../../pages/MasterPersona';
import { STATUS_BADGE, type NpcRow } from '../../types/npc';

interface PersonaPanelProps {
    isActive: boolean;
    onNext: () => void;
}

export const PersonaPanel: React.FC<PersonaPanelProps> = ({ isActive, onNext }) => {
    const [selectedNpc, setSelectedNpc] = useState<NpcRow | null>(NPC_DATA[0]);
    const [isGenerating, setIsGenerating] = useState<boolean>(true);

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>検出されたNPCのペルソナ（性格・口調）を生成します。マスター辞書に存在すればキャッシュを利用します。</span>
            </div>

            <div className="shrink-0">
                <ModelSettings title="ペルソナ生成モデル設定" />
            </div>
            <div className="flex gap-4 flex-1 min-h-0 overflow-hidden relative">
                {/* 左：NPCリスト */}
                <div className="w-1/3 border rounded-xl bg-base-100 flex flex-col min-h-0 overflow-hidden">
                    <ul className="menu w-full bg-base-100 flex-1 overflow-y-auto">
                        <li className="menu-title">NPC一覧 ({NPC_DATA.length})</li>
                        {NPC_DATA.map(npc => (
                            <li key={npc.formId}>
                                <a
                                    className={selectedNpc?.formId === npc.formId ? 'active' : ''}
                                    onClick={() => setSelectedNpc(npc)}
                                >
                                    {npc.name}
                                    <span className={`badge ${STATUS_BADGE[npc.status]} badge-sm ml-auto`}>
                                        {npc.status}
                                    </span>
                                </a>
                            </li>
                        ))}
                    </ul>
                </div>
                {/* 右：ペルソナ詳細 */}
                <div className="w-2/3 flex flex-col min-h-0">
                    <PersonaDetail npc={selectedNpc} />
                </div>

                {isGenerating && (
                    <div className="absolute inset-0 bg-base-100/50 backdrop-blur-[1px] z-10 flex flex-col items-center justify-center gap-4 rounded-xl">
                        <span className="loading loading-spinner text-primary loading-lg"></span>
                        <span className="font-bold text-lg text-base-content/70">ペルソナを自動生成中...</span>
                    </div>
                )}
            </div>

            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0 mt-auto">
                <span className="text-sm font-bold text-gray-500 ml-2">Job: PersonaGeneration ({isGenerating ? 'Running' : 'Stopped'})</span>
                <div className="flex gap-2">
                    <button
                        className={`btn btn-sm ${isGenerating ? 'btn-ghost' : 'btn-outline'}`}
                        onClick={() => setIsGenerating(!isGenerating)}
                    >
                        {isGenerating ? '一時停止' : '再開'}
                    </button>
                    <button className="btn btn-primary btn-sm" onClick={onNext} disabled={isGenerating}>
                        ペルソナを確定して次へ
                    </button>
                </div>
            </div>
        </div>
    );
};
