import React, { useState } from 'react';

const themes = [
    "dark", "light", "cupcake", "bumblebee", "emerald",
    "corporate", "synthwave", "retro", "cyberpunk", "valentine"
];

const ThemeShowcase: React.FC = () => {
    const [theme, setTheme] = useState<string>("dark");

    return (
        <div data-theme={theme} className="min-h-screen bg-base-100 text-base-content p-8 space-y-12 transition-colors duration-200">
            <div className="mb-8 flex flex-col md:flex-row md:items-start justify-between gap-4">
                <div>
                    <h1 className="text-4xl font-bold mb-4">テーマ・デザイン確認用ページ</h1>
                    <p className="opacity-70">
                        現在のDaisyUIテーマ設定の確認用コンポーネント集です。<br />
                        デザインの揺れを排除するため、装飾用Tailwindクラスやインラインスタイルは使用せず、DaisyUIが提供するセマンティックなクラスのみで構築されています。
                    </p>
                </div>

                {/* テーマ切り替えUI */}
                <div className="form-control w-full max-w-xs bg-base-200 p-4 rounded-box">
                    <label className="label">
                        <span className="label-text font-bold">デモ用テーマ切り替え</span>
                    </label>
                    <select
                        className="select select-bordered w-full"
                        value={theme}
                        onChange={(e) => setTheme(e.target.value)}
                    >
                        {themes.map((t) => (
                            <option key={t} value={t}>
                                {t.charAt(0).toUpperCase() + t.slice(1)}
                            </option>
                        ))}
                    </select>
                </div>
            </div>

            {/* 1. ボタン */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">1. ボタン (Buttons)</h2>
                <div className="flex flex-wrap gap-4">
                    <button className="btn">Default</button>
                    <button className="btn btn-primary">Primary</button>
                    <button className="btn btn-secondary">Secondary</button>
                    <button className="btn btn-accent">Accent</button>
                    <button className="btn btn-info">Info</button>
                    <button className="btn btn-success">Success</button>
                    <button className="btn btn-warning">Warning</button>
                    <button className="btn btn-error">Error</button>
                </div>
                <div className="flex flex-wrap gap-4 pt-4">
                    <button className="btn btn-outline">Outline</button>
                    <button className="btn btn-outline btn-primary">Outline Primary</button>
                    <button className="btn btn-ghost">Ghost</button>
                    <button className="btn btn-link">Link</button>
                    <button className="btn btn-active">Active</button>
                    <button className="btn btn-disabled" disabled>Disabled</button>
                </div>
            </section>

            {/* 2. バッジ */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">2. バッジ (Badges)</h2>
                <div className="flex flex-wrap gap-4 items-center">
                    <div className="badge">Default</div>
                    <div className="badge badge-primary">Primary</div>
                    <div className="badge badge-secondary">Secondary</div>
                    <div className="badge badge-accent">Accent</div>
                    <div className="badge badge-outline">Outline</div>
                    <div className="badge badge-info gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="inline-block w-4 h-4 stroke-current"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                        Info
                    </div>
                    <div className="badge badge-success gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="inline-block w-4 h-4 stroke-current"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path></svg>
                        Success
                    </div>
                </div>
            </section>

            {/* 3. カード */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">3. カード (Cards)</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    <div className="card bg-base-100 shadow-xl border border-base-200">
                        <div className="card-body">
                            <h2 className="card-title">デフォルトカード</h2>
                            <p>カードの標準的なデザインです。影と背景色が適用されています。</p>
                            <div className="card-actions justify-end mt-4">
                                <button className="btn btn-primary">アクション</button>
                            </div>
                        </div>
                    </div>

                    <div className="card bg-primary text-primary-content shadow-xl">
                        <div className="card-body">
                            <h2 className="card-title">Primaryカード</h2>
                            <p>Primaryカラーを背景にしたカードです。</p>
                            <div className="card-actions justify-end mt-4">
                                <button className="btn">アクション</button>
                            </div>
                        </div>
                    </div>

                    <div className="card bg-base-100 shadow-xl border border-base-200 card-compact">
                        <div className="card-body">
                            <h2 className="card-title">コンパクトカード</h2>
                            <p>余白が少なく設定されたカードです。</p>
                            <div className="card-actions justify-end mt-4">
                                <button className="btn btn-sm btn-outline">詳細</button>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* 4. アラート */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">4. アラート (Alerts)</h2>
                <div className="space-y-4 max-w-3xl">
                    <div className="alert">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="stroke-info shrink-0 w-6 h-6"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                        <span>新しいアップデートが利用可能です。</span>
                    </div>
                    <div className="alert alert-info">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="stroke-current shrink-0 w-6 h-6"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                        <span>情報メッセージの表示例です。</span>
                    </div>
                    <div className="alert alert-success">
                        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        <span>正常に処理が完了しました。</span>
                    </div>
                    <div className="alert alert-warning">
                        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                        <span>警告: 無効なデータが含まれています。</span>
                    </div>
                    <div className="alert alert-error">
                        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        <span>エラーが発生しました。再度お試しください。</span>
                    </div>
                </div>
            </section>

            {/* 5. 統計情報 (Stats) */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">5. 統計情報 (Stats)</h2>
                <div className="stats shadow border border-base-200 w-full">
                    <div className="stat">
                        <div className="stat-figure text-primary">
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="inline-block w-8 h-8 stroke-current"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path></svg>
                        </div>
                        <div className="stat-title">翻訳総数</div>
                        <div className="stat-value text-primary">25.6K</div>
                        <div className="stat-desc">前月比 +21%</div>
                    </div>

                    <div className="stat">
                        <div className="stat-figure text-secondary">
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="inline-block w-8 h-8 stroke-current"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
                        </div>
                        <div className="stat-title">処理速度</div>
                        <div className="stat-value text-secondary">2.6s</div>
                        <div className="stat-desc">21% 向上</div>
                    </div>

                    <div className="stat">
                        <div className="stat-figure text-secondary">
                            <div className="avatar online">
                                <div className="w-16 rounded-full">
                                    <div className="flex items-center justify-center w-full h-full bg-base-300 font-bold text-xl">AI</div>
                                </div>
                            </div>
                        </div>
                        <div className="stat-value">86%</div>
                        <div className="stat-title">タスク完了率</div>
                        <div className="stat-desc text-secondary">残件 : 3件</div>
                    </div>
                </div>
            </section>

            {/* 6. アコーディオン */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">6. アコーディオン (Accordion)</h2>
                <div className="max-w-2xl bg-base-200 rounded-box">
                    <div className="collapse collapse-arrow bg-base-200">
                        <input type="radio" name="my-accordion-2" defaultChecked />
                        <div className="collapse-title text-xl font-medium">
                            翻訳ルール 1
                        </div>
                        <div className="collapse-content">
                            <p>このセクションには翻訳ルールの詳細が表示されます。</p>
                        </div>
                    </div>
                    <div className="collapse collapse-arrow bg-base-200">
                        <input type="radio" name="my-accordion-2" />
                        <div className="collapse-title text-xl font-medium">
                            翻訳ルール 2
                        </div>
                        <div className="collapse-content">
                            <p>常に丁寧語で出力する設定となっています。</p>
                        </div>
                    </div>
                    <div className="collapse collapse-arrow bg-base-200">
                        <input type="radio" name="my-accordion-2" />
                        <div className="collapse-title text-xl font-medium">
                            除外用語
                        </div>
                        <div className="collapse-content">
                            <p>禁止リストに含まれる用語の設定です。</p>
                        </div>
                    </div>
                </div>
            </section>

            {/* 7. タブ */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">7. タブ (Tabs)</h2>
                <div role="tablist" className="tabs tabs-bordered w-full max-w-xl">
                    <a role="tab" className="tab">設定</a>
                    <a role="tab" className="tab tab-active">履歴</a>
                    <a role="tab" className="tab">統計</a>
                </div>
                <div role="tablist" className="tabs tabs-lifted w-full max-w-xl mt-4">
                    <a role="tab" className="tab">設定</a>
                    <a role="tab" className="tab tab-active">履歴</a>
                    <a role="tab" className="tab">統計</a>
                </div>
                <div role="tablist" className="tabs tabs-boxed w-full max-w-xl mt-4 inline-flex">
                    <a role="tab" className="tab">設定</a>
                    <a role="tab" className="tab tab-active bg-primary text-primary-content">履歴</a>
                    <a role="tab" className="tab">統計</a>
                </div>
            </section>

            {/* 8. フォーム */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">8. フォーム (Forms)</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl">
                    <div className="space-y-4">
                        <div className="form-control w-full">
                            <label className="label">
                                <span className="label-text">テキスト入力</span>
                            </label>
                            <input type="text" placeholder="文字を入力..." className="input input-bordered w-full" />
                        </div>
                        <div className="form-control w-full">
                            <label className="label">
                                <span className="label-text">セレクトボックス</span>
                            </label>
                            <select className="select select-bordered w-full">
                                <option disabled selected>選択してください</option>
                                <option>オプション 1</option>
                                <option>オプション 2</option>
                            </select>
                        </div>
                        <div className="form-control w-full">
                            <label className="label">
                                <span className="label-text">テキストエリア</span>
                            </label>
                            <textarea className="textarea textarea-bordered h-24 w-full" placeholder="詳細を入力..."></textarea>
                        </div>
                    </div>
                    <div className="space-y-4 p-4 border border-base-300 rounded-box">
                        <div className="form-control">
                            <label className="label cursor-pointer">
                                <span className="label-text">チェックボックス (Primary)</span>
                                <input type="checkbox" defaultChecked className="checkbox checkbox-primary" />
                            </label>
                        </div>
                        <div className="form-control">
                            <label className="label cursor-pointer">
                                <span className="label-text">トグルスイッチ (Secondary)</span>
                                <input type="checkbox" className="toggle toggle-secondary" defaultChecked />
                            </label>
                        </div>
                        <div className="form-control">
                            <label className="label cursor-pointer">
                                <span className="label-text">ラジオボタン A</span>
                                <input type="radio" name="radio-10" className="radio radio-accent" defaultChecked />
                            </label>
                        </div>
                        <div className="form-control">
                            <label className="label cursor-pointer">
                                <span className="label-text">ラジオボタン B</span>
                                <input type="radio" name="radio-10" className="radio radio-accent" />
                            </label>
                        </div>
                        <div className="form-control">
                            <label className="label">
                                <span className="label-text">レンジスライダー</span>
                            </label>
                            <input type="range" min="0" max="100" defaultValue="40" className="range range-primary" />
                        </div>
                    </div>
                </div>
            </section>

            {/* 9. プログレス / ロード状態 */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">9. プログレス・ロード (Progress / Loading)</h2>
                <div className="flex flex-col gap-6 max-w-2xl">
                    <div className="flex gap-4 items-center">
                        <span className="w-24">Loading:</span>
                        <span className="loading loading-spinner text-primary"></span>
                        <span className="loading loading-dots text-secondary"></span>
                        <span className="loading loading-ring text-accent"></span>
                    </div>
                    <div className="flex flex-col gap-2">
                        <span>Progress (Primary):</span>
                        <progress className="progress progress-primary w-full" value="40" max="100"></progress>
                    </div>
                    <div className="flex flex-col gap-2">
                        <span>Progress (Indeterminate):</span>
                        <progress className="progress w-full"></progress>
                    </div>
                </div>
            </section>

            {/* 10. テーブル */}
            <section className="space-y-4">
                <h2 className="text-2xl font-semibold border-b border-base-300 pb-2">10. テーブル (Table)</h2>
                <div className="overflow-x-auto border border-base-200 rounded-box">
                    <table className="table table-zebra w-full">
                        <thead>
                            <tr>
                                <th></th>
                                <th>名前</th>
                                <th>役割</th>
                                <th>ステータス</th>
                                <th>アクション</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <th>1</th>
                                <td>Cy Ganderton</td>
                                <td>Quality Control Specialist</td>
                                <td><div className="badge badge-success badge-sm">Active</div></td>
                                <td><button className="btn btn-xs">詳細</button></td>
                            </tr>
                            <tr>
                                <th>2</th>
                                <td>Hart Hagerty</td>
                                <td>Desktop Support Technician</td>
                                <td><div className="badge badge-ghost badge-sm">Inactive</div></td>
                                <td><button className="btn btn-xs">詳細</button></td>
                            </tr>
                            <tr>
                                <th>3</th>
                                <td>Brice Swyre</td>
                                <td>Tax Accountant</td>
                                <td><div className="badge badge-error badge-sm">Error</div></td>
                                <td><button className="btn btn-xs">詳細</button></td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>

        </div>
    );
};

export default ThemeShowcase;
