import { create } from 'zustand'
import { FrontendTask } from '../types/task'
import * as Events from '../wailsjs/runtime/runtime'
import { GetActiveTasks, GetAllTasks, ResumeTask, CancelTask } from '../wailsjs/go/task/Bridge'

interface TaskState {
    tasks: Record<string, FrontendTask>;
    isLoading: boolean;

    setTasks: (tasks: FrontendTask[]) => void;
    updateTask: (task: FrontendTask) => void;
    removeTask: (taskId: string) => void;

    fetchActiveTasks: () => Promise<void>;
    resumeTask: (taskId: string) => Promise<void>;
    cancelTask: (taskId: string) => Promise<void>;
}

export const useTaskStore = create<TaskState>((set, get) => ({
    tasks: {},
    isLoading: false,

    setTasks: (tasks) => {
        const taskMap: Record<string, FrontendTask> = {};
        tasks.forEach(t => {
            taskMap[t.id] = t;
        });
        set({ tasks: taskMap });
    },

    updateTask: (task) => {
        if (task.status === 'completed') {
            get().removeTask(task.id);
            return;
        }
        set((state) => ({
            tasks: {
                ...state.tasks,
                [task.id]: task
            }
        }));
    },

    removeTask: (taskId) => {
        set((state) => {
            const newTasks = { ...state.tasks };
            delete newTasks[taskId];
            return { tasks: newTasks };
        });
    },

    fetchActiveTasks: async () => {
        if (typeof GetAllTasks !== 'function' && typeof GetActiveTasks !== 'function') {
            console.warn('Wails task bindings are not available');
            return;
        }
        set({ isLoading: true });
        try {
            const tasks = typeof GetActiveTasks === 'function'
                ? (await GetActiveTasks() as any) as FrontendTask[]
                : (await GetAllTasks() as any) as FrontendTask[];
            const activeOnly = (tasks || []).filter((t) => t.status !== 'completed');
            get().setTasks(activeOnly);
        } catch (error) {
            console.error('Failed to fetch tasks:', error);
        } finally {
            set({ isLoading: false });
        }
    },

    resumeTask: async (taskId) => {
        try {
            await ResumeTask(taskId);
        } catch (error) {
            console.error('Failed to resume task:', error);
            throw error;
        }
    },

    cancelTask: async (taskId) => {
        try {
            await CancelTask(taskId);
        } catch (error) {
            console.error('Failed to cancel task:', error);
            throw error;
        }
    }
}));

let listenersInitialized = false;

// Initialize event listeners
export const initTaskListeners = () => {
    if (listenersInitialized) return;

    if (typeof Events.EventsOn !== 'function') {
        console.warn('Wails EventsOn is not available');
        return;
    }

    Events.EventsOn('task:updated', (task: FrontendTask) => {
        useTaskStore.getState().updateTask(task);
    });

    Events.EventsOn('task:phase_completed', (payload: { taskId: string, phase: string, summary: any }) => {
        console.log('Phase completed:', payload);
        // Additional handling can be added here (e.g., refetching data related to the task)
    });

    listenersInitialized = true;
};
