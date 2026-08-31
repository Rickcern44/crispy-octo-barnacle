import roadmap from "./generated/roadmap.json";

export type Roadmap = typeof roadmap;
export type Item = Roadmap["items"][number];

const planItemIDs = new Map(
	roadmap.plans.map((plan) => [plan.id, plan.item_id]),
);
const completedDates = new Map<number, string>();

for (const task of roadmap.completed_tasks) {
	const itemID = planItemIDs.get(task.plan_id);
	if (itemID && task.completed_at) {
		const date = task.completed_at.slice(0, 10);
		const previousDate = completedDates.get(itemID);
		if (!previousDate || date > previousDate) completedDates.set(itemID, date);
	}
}

export function milestoneDate(item: Item) {
	return item.status === "Done"
		? (completedDates.get(item.id) ?? item.target_date)
		: item.target_date;
}

export function milestoneKind(item: Item) {
	if (!milestoneDate(item)) return "unscheduled";
	return item.status === "Done" ? "done" : "planned";
}

export const milestones = [...roadmap.items].sort((left, right) => {
	const leftDate = milestoneDate(left);
	const rightDate = milestoneDate(right);
	if (!leftDate) return 1;
	if (!rightDate) return -1;
	return leftDate.localeCompare(rightDate);
});

export function plansForItem(itemID: number) {
	return roadmap.plans
		.filter((plan) => plan.item_id === itemID)
		.map((plan) => ({
			...plan,
			tasks: roadmap.tasks.filter((task) => task.plan_id === plan.id),
		}));
}

export function reportsForItem(itemID: number) {
	return roadmap.reports.filter((report) => report.item_id === itemID);
}

export function lifecycleForItem(itemID: number) {
	return {
		phases: roadmap.phases.filter((phase) => phase.ItemID === itemID),
		criteria: roadmap.criteria.filter(
			(criterion) => criterion.ItemID === itemID,
		),
	};
}

export { roadmap };
