import {create} from 'zustand'
import {FrontendTask} from '../types/task'
import * as Events from '../wailsjs/runtime/runtime'
import {CancelTask, DeleteTask, GetActiveTasks, GetAllTasks, ResumeTask} from '../wailsjs/go/controller/TaskController'
import {
    CancelTask as CancelPersonaTask,
    ResumeTask as ResumePersonaTask
} from '../wailsjs/go/controller/PersonaTaskController'

interface TaskState {
    tasks: Record<string, FrontendTask>;
    isLoading: boolean;

    setTasks: (tasks: FrontendTask[]) => void;
    updateTask: (task: FrontendTask) => void;
    removeTask: (taskId: string) => void;

    fetchActiveTasks: () => Promise<void>;
    resumeTask: (task: FrontendTask) => Promise<void>;
    cancelTask: (task: FrontendTask) => Promise<void>;
    deleteTask: (task: FrontendTask) => Promise<void>;
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
                ? (await GetActiveTasks() as unknown) as FrontendTask[]
                : (await GetAllTasks() as unknown) as FrontendTask[];
            const activeOnly = (tasks || []).filter((t) => t.status !== 'completed');
            get().setTasks(activeOnly);
        } catch (error) {
            console.error('Failed to fetch tasks:', error);
        } finally {
            set({ isLoading: false });
        }
    },

    resumeTask: async (task) => {
        try {
            if (task.type === 'persona_extraction') {
                await ResumePersonaTask(task.id);
                return;
            }
            await ResumeTask(task.id);
        } catch (error) {
            console.error('Failed to resume task:', error);
            throw error;
        }
    },

    cancelTask: async (task) => {
        try {
            if (task.type === 'persona_extraction') {
                await CancelPersonaTask(task.id);
                return;
            }
            await CancelTask(task.id);
        } catch (error) {
            console.error('Failed to cancel task:', error);
            throw error;
        }
    },

    deleteTask: async (task) => {
        try {
            await DeleteTask(task.id);
            get().removeTask(task.id);
        } catch (error) {
            console.error('Failed to delete task:', error);
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

    Events.EventsOn('task:phase_completed', (payload: { taskId: string, phase: string, summary: unknown }) => {
        console.log('Phase completed:', payload);
        // Additional handling can be added here (e.g., refetching data related to the task)
    });

    listenersInitialized = true;
};

