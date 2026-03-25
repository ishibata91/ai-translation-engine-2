import {useDashboard} from '../hooks/features/dashboard/useDashboard';

/**
 * 実行中および再開可能なジョブ一覧を表示する。
 */
export default function Dashboard() {
    const {
        sortedTasks,
        cancelTask,
        handleTaskClick,
        handleResumeClick,
        isTaskManageable,
        deleteTargetTask,
        isDeleteModalOpen,
        isDeleting,
        deleteError,
        handleDeleteClick,
        handleCancelDelete,
        handleConfirmDelete,
    } = useDashboard();

    return (
        <div className="flex flex-col w-full p-4 gap-4">
            <div className="navbar bg-base-100 rounded-box shadow-sm px-4">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">ダッシュボード (Dashboard)</span>
                </div>
            </div>

            <div className="card bg-base-100 shadow-sm border border-base-200 flex-1">
                <div className="card-body">
                    <h2 className="card-title text-base">ジョブ一覧</h2>
                    <div className="overflow-x-auto">
                        <table className="table w-full">
                            <thead>
                                <tr>
                                    <th>ジョブ名</th>
                                    <th>タイプ</th>
                                    <th>フェーズ</th>
                                    <th>進捗</th>
                                    <th>ステータス</th>
                                    <th>アクション</th>
                                </tr>
                            </thead>
                            <tbody>
                                {sortedTasks.length === 0 ? (
                                    <tr>
                                        <td colSpan={6} className="text-center text-base-content/50 py-4">
                                            ジョブはありません
                                        </td>
                                    </tr>
                                ) : (
                                    sortedTasks.map((task) => (
                                        <tr key={task.id} className="hover cursor-pointer" onClick={() => handleTaskClick(task)}>
                                            <td>{task.name}</td>
                                            <td>
                                                <div className="badge badge-ghost badge-sm">{task.type}</div>
                                            </td>
                                            <td>{task.phase}</td>
                                            <td>
                                                <div className="flex items-center gap-2">
                                                    <progress
                                                        className={`progress w-20 ${task.status === 'failed' ? 'progress-error' : 'progress-primary'}`}
                                                        value={task.progress}
                                                        max="100"
                                                    ></progress>
                                                    <span className="text-xs font-mono">{task.progress.toFixed(1)}%</span>
                                                </div>
                                            </td>
                                            <td>
                                                <div
                                                    className={`badge badge-sm ${
                                                        task.status === 'running'
                                                            ? 'badge-primary'
                                                            : task.status === 'paused' || task.status === 'request_generated'
                                                              ? 'badge-warning'
                                                              : task.status === 'failed' || task.status === 'cancelled'
                                                                ? 'badge-error'
                                                                : task.status === 'completed'
                                                                  ? 'badge-success'
                                                                  : 'badge-ghost'
                                                    }`}
                                                >
                                                    {task.status}
                                                </div>
                                            </td>
                                            <td onClick={(e) => e.stopPropagation()}>
                                                <div className="flex gap-2">
                                                    {(task.status === 'paused' ||
                                                        task.status === 'request_generated' ||
                                                        task.status === 'failed' ||
                                                        task.status === 'cancelled' ||
                                                        task.status === 'pending') && (
                                                        <button className="btn btn-xs btn-success" onClick={() => handleResumeClick(task)}>
                                                            再開
                                                        </button>
                                                    )}
                                                    {task.status === 'running' && (
                                                        <button className="btn btn-xs btn-outline btn-error" onClick={() => cancelTask(task)}>
                                                            停止
                                                        </button>
                                                    )}
                                                    {isTaskManageable(task) && (
                                                        <button
                                                            className="btn btn-xs btn-outline btn-error"
                                                            data-testid={`dashboard-delete-${task.id}`}
                                                            onClick={() => handleDeleteClick(task)}
                                                        >
                                                            削除
                                                        </button>
                                                    )}
                                                </div>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            <dialog className={`modal ${isDeleteModalOpen ? 'modal-open' : ''}`}>
                <div className="modal-box" data-testid="dashboard-delete-modal">
                    <h3 className="font-bold text-lg">タスクを削除しますか？</h3>
                    <p className="py-3 text-sm text-base-content/80">
                        {deleteTargetTask?.name ?? '対象タスク'} を削除します。task 本体と task 管理に紐づく中間成果物が削除されます。
                    </p>
                    <p className="text-xs text-base-content/60">artifact 正本や共有成果物は削除されません。</p>
                    {deleteError != null && <p className="text-sm text-error mt-3">{deleteError}</p>}
                    <div className="modal-action">
                        <button className="btn btn-ghost" disabled={isDeleting} onClick={handleCancelDelete}>
                            キャンセル
                        </button>
                        <button
                            className="btn btn-error"
                            data-testid="dashboard-delete-confirm"
                            disabled={isDeleting}
                            onClick={() => {
                                void handleConfirmDelete();
                            }}
                        >
                            {isDeleting ? '削除中...' : '削除する'}
                        </button>
                    </div>
                </div>
                <form className="modal-backdrop" method="dialog">
                    <button type="button" onClick={handleCancelDelete}>
                        close
                    </button>
                </form>
            </dialog>
        </div>
    );
}
