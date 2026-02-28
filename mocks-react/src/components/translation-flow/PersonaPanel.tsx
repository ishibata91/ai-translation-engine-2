import React from 'react';

interface PersonaPanelProps {
    isActive: boolean;
    onNext: () => void;
}

export const PersonaPanel: React.FC<PersonaPanelProps> = ({ isActive, onNext }) => {
    return (
        <div className={`tab-content-panel flex-col gap-4 h-full ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>検出されたNPCのペルソナ（性格・口調）を生成します。マスター辞書に存在すればキャッシュを利用します。</span>
                <div className="flex-none">
                    <button className="btn btn-sm btn-primary">LLMで一括生成</button>
                </div>
            </div>
            <div className="flex gap-4 flex-1 min-h-0 overflow-hidden">
                {/* 左：NPCリスト */}
                <div className="w-1/3 border rounded-xl bg-base-100 overflow-y-auto">
                    <ul className="menu w-full bg-base-100">
                        <li className="menu-title">NPC一覧 (24)</li>
                        <li><a className="active">Ulfric Stormcloak <span className="badge badge-success badge-sm ml-auto">完了</span></a></li>
                        <li><a>General Tullius <span className="badge badge-success badge-sm ml-auto">完了</span></a></li>
                        <li><a>Elisif the Fair <span className="badge badge-warning badge-sm ml-auto">生成中</span></a></li>
                        <li><a>Whiterun Guard <span className="badge badge-ghost badge-sm ml-auto">未生成</span></a></li>
                    </ul>
                </div>
                {/* 右：ペルソナ詳細 */}
                <div className="w-2/3 border rounded-xl bg-base-100 p-4 flex flex-col gap-4 overflow-y-auto">
                    <h3 className="text-xl font-bold border-b pb-2">Ulfric Stormcloak (0001414D)</h3>
                    <div className="flex gap-2">
                        <span className="badge badge-outline">Race: Nord</span>
                        <span className="badge badge-outline">Sex: Male</span>
                        <span className="badge badge-outline">Class: Warrior</span>
                    </div>
                    <div className="form-control flex-1 flex flex-col min-h-0">
                        <label className="label"><span className="label-text font-bold">生成されたペルソナ情報 (プロンプト動的注入用)</span></label>
                        <textarea className="textarea textarea-bordered flex-1 text-base leading-relaxed" defaultValue={`誇り高く、カリスマ性のあるストームクロークの反乱軍リーダー。雄弁で威厳のある口調。\n一人称: 私、俺\n二人称: お前、貴様\n特徴: スカイリムの独立とノルドの誇りを強調する。「～だ」「～だろう」「～なのだ」といった断定的で力強い語尾を多用する。若者には厳しくも期待を込めて接する。`}></textarea>
                    </div>
                    <div className="flex justify-end gap-2 mt-2">
                        <button className="btn btn-outline btn-sm">再生成</button>
                        <button className="btn btn-secondary btn-sm">保存</button>
                    </div>
                </div>
            </div>
            <div className="flex justify-end gap-2 shrink-0">
                <button className="btn btn-primary" onClick={onNext}>ペルソナを確定して次へ</button>
            </div>
        </div>
    );
};
