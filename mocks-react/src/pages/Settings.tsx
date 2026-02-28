import React from 'react';

const Settings: React.FC = () => {
    return (
        <div className="flex flex-col w-full p-4 gap-4 pb-20">
            {/* ヘッダー部分 */}
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">設定 (Settings)</span>
                    <div className="flex gap-4">
                        <button className="btn btn-ghost btn-sm">リセット</button>
                        <button className="btn btn-primary btn-sm">設定を保存</button>
                    </div>
                </div>
            </div>

            {/* コンテンツエリア */}
            <div className="grid grid-cols-1 gap-4">

                {/* LLMプロバイダ設定 */}
                <div className="card bg-base-100 border border-base-200 shadow-sm">
                    <div className="card-body">
                        <h2 className="card-title text-lg border-b border-base-200 pb-2">LLMプロバイダ設定</h2>

                        <div className="flex flex-col gap-4 mt-2">
                            <div className="flex flex-col gap-1">
                                <label className="label pb-0">
                                    <span className="label-text font-bold">アクティブなプロバイダ</span>
                                </label>
                                <select className="select select-bordered w-full max-w-md">
                                    <option>Google Gemini</option>
                                    <option>OpenAI (GPT-4o)</option>
                                    <option>xAI (Grok)</option>
                                    <option>Local LLM (llama.cpp / Ollama)</option>
                                </select>
                            </div>

                            <div className="flex flex-col gap-1">
                                <label className="label pb-0">
                                    <span className="label-text font-bold">Gemini APIキー</span>
                                </label>
                                <input type="password" placeholder="AIzaSy..." className="input input-bordered w-full max-w-md" />
                            </div>

                            <div className="flex flex-col gap-1">
                                <label className="label pb-0">
                                    <span className="label-text font-bold">OpenAI APIキー</span>
                                </label>
                                <input type="password" placeholder="sk-..." className="input input-bordered w-full max-w-md" />
                            </div>

                            <div className="flex flex-col gap-1">
                                <label className="label pb-0">
                                    <span className="label-text font-bold">xAI APIキー</span>
                                </label>
                                <input type="password" placeholder="xai-..." className="input input-bordered w-full max-w-md" />
                            </div>
                        </div>
                    </div>
                </div>

                {/* RAGプロンプト・テンプレート管理 */}
                <div className="card bg-base-100 border border-base-200 shadow-sm">
                    <div className="card-body">
                        <h2 className="card-title text-lg border-b border-base-200 pb-2">RAGプロンプト・テンプレート管理</h2>
                        <div className="text-sm text-base-content/70 mt-1">
                            modules/prompts.py で定義されているシステム全体のRAG構築・翻訳推論に用いる指示を調整します。
                        </div>

                        <div className="flex flex-col gap-2 mt-4">
                            {/* セクション1: 共通システム構築 */}
                            <div className="collapse collapse-arrow bg-base-200 border border-base-300">
                                <input type="radio" name="prompt-accordion" defaultChecked />
                                <div className="collapse-title font-bold text-base">1. 共通システムプロンプト (System Prompts)</div>
                                <div className="collapse-content flex flex-col gap-4 pt-4 border-t border-base-300 bg-base-100">
                                    <div className="flex flex-col gap-1">
                                        <label className="label pb-0"><span className="label-text font-bold">ベース指示 (PROMPT_BASE)</span></label>
                                        <textarea className="textarea textarea-bordered h-24 font-mono text-sm leading-relaxed"
                                            defaultValue={`You are a localization specialist for 'The Elder Scrolls V: Skyrim'. Translate the text inside <source_text> into natural Japanese suitable for a fantasy RPG setting.`}></textarea>
                                    </div>
                                    <div className="flex flex-col gap-1">
                                        <label className="label pb-0"><span className="label-text font-bold text-error">制約事項 (PROMPT_PROHIBITION)</span></label>
                                        <textarea className="textarea textarea-bordered h-32 font-mono text-sm leading-relaxed"
                                            defaultValue={`- STRICTLY PROHIBITED: Do not add tags or placeholders that do not exist in the original text.\n- STRICTLY PROHIBITED: Do not incorporate factual details from <summary> into the translation text. Use <summary> ONLY for context understanding.`}></textarea>
                                    </div>
                                </div>
                            </div>

                            {/* セクション2: レコード別タスク指示 */}
                            <div className="collapse collapse-arrow bg-base-200 border border-base-300">
                                <input type="radio" name="prompt-accordion" />
                                <div className="collapse-title font-bold text-base">2. レコード種別固有指示 (Record Specific Instructions)</div>
                                <div className="collapse-content flex flex-col gap-4 pt-4 border-t border-base-300 bg-base-100">
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div className="flex flex-col gap-1">
                                            <label className="label pb-0"><span className="label-text font-mono font-bold">WEAP (武器)</span></label>
                                            <textarea className="textarea textarea-bordered min-h-24 font-mono text-sm"
                                                defaultValue={`Target is a weapon name or description. Convey material, magical effects, and weightiness.`}></textarea>
                                        </div>
                                        <div className="flex flex-col gap-1">
                                            <label className="label pb-0"><span className="label-text font-mono font-bold">INFO (会話)</span></label>
                                            <textarea className="textarea textarea-bordered min-h-24 font-mono text-sm"
                                                defaultValue={`Target is an NPC dialogue. Translate into natural spoken Japanese based on <meta> speaker attributes.`}></textarea>
                                        </div>
                                    </div>
                                    <div className="flex justify-end mt-2">
                                        <button className="btn btn-ghost btn-sm">一覧をすべて編集...</button>
                                    </div>
                                </div>
                            </div>

                            {/* セクション3: 話法・トーンマッピング */}
                            <div className="collapse collapse-arrow bg-base-200 border border-base-300">
                                <input type="radio" name="prompt-accordion" />
                                <div className="collapse-title font-bold text-base">3. トーン＆パーソナリティ紐付け (Tone Mapping)</div>
                                <div className="collapse-content flex flex-col gap-4 pt-4 border-t border-base-300 bg-base-100">
                                    <div className="tabs tabs-boxed bg-base-200 w-fit">
                                        <a className="tab tab-active">種族 (Race_Tone)</a>
                                        <a className="tab">ボイス (Voice_Tone)</a>
                                    </div>
                                    <div className="flex flex-col gap-4 mt-2">
                                        <div className="flex flex-col gap-1">
                                            <label className="label pb-0"><span className="label-text font-mono font-bold text-primary">Khajiit / KhajiitRace</span></label>
                                            <textarea className="textarea textarea-bordered h-24 font-mono text-sm"
                                                defaultValue={`[STRICT RULE] Always refer to self in the third person. First-person pronouns are STRICTLY FORBIDDEN. Use distinctive Khajiit speech patterns.`}></textarea>
                                        </div>
                                        <div className="flex flex-col gap-1">
                                            <label className="label pb-0"><span className="label-text font-mono font-bold text-primary">HighElf / Altmer</span></label>
                                            <textarea className="textarea textarea-bordered h-24 font-mono text-sm"
                                                defaultValue={`Arrogant and politely rude elite tone. Show off superior intelligence and culture.`}></textarea>
                                        </div>
                                    </div>
                                    <div className="flex justify-end mt-2">
                                        <button className="btn btn-ghost btn-sm">全マッピングを編集...</button>
                                    </div>
                                </div>
                            </div>

                        </div>
                    </div>
                </div>

            </div>
        </div>
    );
};

export default Settings;
