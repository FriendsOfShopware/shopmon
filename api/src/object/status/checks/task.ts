import type { ScheduledTask } from '../../../types/index.ts';
import type { Checker, CheckerInput, CheckerOutput } from '../registery.ts';

export default class implements Checker {
    async check(input: CheckerInput, result: CheckerOutput): Promise<void> {
        for (const task of input.scheduledTasks) {
            if (isTaskOverdue(task) && task.interval > 3600) {
                result.warning(
                    `task.${task.name}`,
                    `Task ${task.name} is overdue`,
                );
            }
        }

        result.success('task.all', 'All scheduled tasks are running correctly');
    }
}

function isTaskOverdue(task: ScheduledTask) {
    // Some timestamps are always UTC
    const currentTime = new Date();
    const nextExecute = new Date(task.nextExecutionTime);

    return currentTime.getTime() > nextExecute.getTime();
}
