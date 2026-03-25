import {useMemo, useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {useShallow} from 'zustand/react/shallow';
import {useTaskStore} from '../../../store/taskStore';
import type {FrontendTask, TaskStatus} from '../../../types/task';

const manageableStatuses: ReadonlySet<TaskStatus> = new Set(['pending', 'paused', 'request_generated', 'failed', 'cancelled']);

const formatDeleteError = (error: unknown): string => {
    if (error instanceof Error && error.message !== '') {
        return error.message;
    }
    return 'タスクの削除に失敗しました。時間をおいて再試行してください。';
};

/**
 * ダッシュボード画面で必要なジョブ一覧 state と操作を返す。
 */
export function useDashboard() {
    const tasks = useTaskStore(useShallow((state) => Object.values(state.tasks)));
    const resumeTask = useTaskStore((state) => state.resumeTask);
    const cancelTask = useTaskStore((state) => state.cancelTask);
    const deleteTask = useTaskStore((state) => state.deleteTask);
    const navigate = useNavigate();

    const [deleteTargetTask, setDeleteTargetTask] = useState<FrontendTask | null>(null);
    const [isDeleting, setIsDeleting] = useState(false);
    const [deleteError, setDeleteError] = useState<string | null>(null);

    const sortedTasks = useMemo(
        () => [...tasks].sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()),
        [tasks],
    );

    const handleTaskClick = (task: FrontendTask) => {
        const navigationState = { state: { taskId: task.id, phase: task.phase } };

        switch (task.type) {
            case 'dictionary_build':
                navigate('/dictionary', navigationState);
                return;
            case 'persona_extraction':
                navigate('/master_persona', navigationState);
                return;
            case 'translation_project':
                navigate('/translation_flow', navigationState);
                return;
        }
    };

    const handleResumeClick = (task: FrontendTask) => {
        if (task.type === 'persona_extraction') {
            navigate('/master_persona', { state: { taskId: task.id, phase: task.phase, resumeFromDashboard: true } });
            return;
        }

        void resumeTask(task);
    };

    const isTaskManageable = (task: FrontendTask): boolean => manageableStatuses.has(task.status);

    const handleDeleteClick = (task: FrontendTask) => {
        setDeleteError(null);
        setDeleteTargetTask(task);
    };

    const handleCancelDelete = () => {
        if (isDeleting) {
            return;
        }
        setDeleteError(null);
        setDeleteTargetTask(null);
    };

    const handleConfirmDelete = async () => {
        if (deleteTargetTask == null || isDeleting) {
            return;
        }

        setIsDeleting(true);
        setDeleteError(null);
        try {
            await deleteTask(deleteTargetTask);
            setDeleteTargetTask(null);
        } catch (error) {
            setDeleteError(formatDeleteError(error));
        } finally {
            setIsDeleting(false);
        }
    };

    return {
        sortedTasks,
        cancelTask,
        handleTaskClick,
        handleResumeClick,
        isTaskManageable,
        deleteTargetTask,
        isDeleteModalOpen: deleteTargetTask !== null,
        isDeleting,
        deleteError,
        handleDeleteClick,
        handleCancelDelete,
        handleConfirmDelete,
    };
}
