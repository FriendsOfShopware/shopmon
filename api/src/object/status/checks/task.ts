import { ScheduledTask } from "../../../../../shared/shop";
import { Checker, CheckerInput, CheckerOutput } from "../registery";

export default class implements Checker {
    async check(input: CheckerInput, result: CheckerOutput): Promise<void> {
        for (let task of input.scheduledTasks) {
            if (isTaskOverdue(task) && task.interval > 3600) {
                result.warning(`Task ${task.name} is overdue`);
            }
        }
    }
}

function isTaskOverdue(task: ScheduledTask) {
    // Some timestamps are always UTC
    var currentTime = new Date();
    const nextExecute = new Date(task.nextExecutionTime);
  
    return currentTime.getTime() > nextExecute.getTime();
  }