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

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>検出されたNPCのペルソナ（性格・口調）を生成します。マスター辞書に存在すればキャッシュを利用します。</span>
                <div className="flex-none">
                    <button className="btn btn-sm btn-primary">LLMで一括生成</button>
                </div>
            </div>

            <div className="shrink-0">
                <ModelSettings title="ペルソナ生成モデル設定" />
            </div>
            <div className="flex gap-4 flex-1 min-h-0 overflow-hidden">
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
            </div>
            <div className="flex justify-end gap-2 shrink-0">
                <button className="btn btn-primary" onClick={onNext}>ペルソナを確定して次へ</button>
            </div>
        </div>
    );
};
