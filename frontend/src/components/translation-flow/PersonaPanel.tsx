import {useState} from 'react';

interface PersonaPanelProps {
    isActive: boolean;
    onNext: () => void;
}

interface NpcRow {
    formId: string;
    name: string;
}

const NPC_DATA: NpcRow[] = [
    { formId: '00013B9B', name: 'UlfricStormcloak' },
    { formId: '000198BB', name: 'Ralof' },
    { formId: '0001326D', name: 'Hadvar' },
];

function PlaceholderModelSettings({ title }: { title: string }) {
    return (
        <div className="p-4 border border-dashed border-base-300 text-center text-gray-500 rounded-xl">
            [{title} プレースホルダー]
        </div>
    );
}

function PlaceholderPersonaDetail({ npc }: { npc: NpcRow | null }) {
    if (npc === null) {
        return <div className="p-4 text-center text-gray-400">NPC を選択してください</div>;
    }

    return (
        <div className="p-4 flex flex-col gap-2">
            <span className="font-bold">{npc.name}</span>
            <span className="text-sm text-base-content/70">FormID: {npc.formId}</span>
            <span className="text-sm text-base-content/70">Persona Detail Placeholder</span>
        </div>
    );
}

export function PersonaPanel({ isActive, onNext }: PersonaPanelProps) {
    const [selectedNpc, setSelectedNpc] = useState<NpcRow | null>(NPC_DATA[0] ?? null);
    const [isGenerating, setIsGenerating] = useState(true);

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>検出されたNPCのペルソナ（性格・口調）を生成します。マスター辞書に存在すればキャッシュを利用します。</span>
            </div>

            <div className="shrink-0">
                <PlaceholderModelSettings title="ペルソナ生成モデル設定" />
            </div>

            <div className="flex gap-4 flex-1 min-h-0 overflow-hidden relative">
                <div className="w-1/3 border rounded-xl bg-base-100 flex flex-col min-h-0 overflow-hidden">
                    <ul className="menu w-full bg-base-100 flex-1 overflow-y-auto">
                        <li className="menu-title">NPC一覧 ({NPC_DATA.length})</li>
                        {NPC_DATA.map((npc) => (
                            <li key={npc.formId}>
                                <button
                                    type="button"
                                    className={selectedNpc?.formId === npc.formId ? 'active text-left' : 'text-left'}
                                    onClick={() => setSelectedNpc(npc)}
                                >
                                    {npc.name}
                                </button>
                            </li>
                        ))}
                    </ul>
                </div>

                <div className="w-2/3 flex flex-col min-h-0 rounded-xl border bg-base-100">
                    <PlaceholderPersonaDetail npc={selectedNpc} />
                </div>

                {isGenerating && (
                    <div className="absolute inset-0 bg-base-100/50 backdrop-blur-[1px] z-10 flex flex-col items-center justify-center gap-4 rounded-xl">
                        <span className="loading loading-spinner text-primary loading-lg"></span>
                        <span className="font-bold text-lg text-base-content/70">ペルソナを自動生成中...</span>
                    </div>
                )}
            </div>

            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0 mt-auto">
                <span className="text-sm font-bold text-gray-500 ml-2">
                    Job: PersonaGeneration ({isGenerating ? 'Running' : 'Stopped'})
                </span>
                <div className="flex gap-2">
                    <button
                        type="button"
                        className={`btn btn-sm ${isGenerating ? 'btn-ghost' : 'btn-outline'}`}
                        onClick={() => setIsGenerating(!isGenerating)}
                    >
                        {isGenerating ? '一時停止' : '再開'}
                    </button>
                    <button type="button" className="btn btn-primary btn-sm" onClick={onNext} disabled={isGenerating}>
                        ペルソナを確定して次へ
                    </button>
                </div>
            </div>
        </div>
    );
}
