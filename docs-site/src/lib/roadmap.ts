import roadmap from "./generated/roadmap.json";

export type Roadmap = typeof roadmap;
export type Item = Roadmap["items"][number];
export type Specification = {
	title: string;
	description: string;
	done: boolean;
};
export type FeatureRelationship = {
	id: number;
	source_item_id: number;
	source_item_title: string;
	target_item_id: number;
	target_item_title: string;
	relationship_type: "extends" | "depends_on" | "replaces";
	created_at: string;
};

const relationships: FeatureRelationship[] = roadmap.relationships;

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

// Lifecycle completion is authoritative for presentation. Historical metadata
// may retain a lower progress value after an item has been delivered.
export function displayProgress(item: Item) {
	return item.status === "Done" ? 100 : item.progress;
}

export function specificationsForItem(item: Item): Specification[] {
	try {
		const specifications = JSON.parse(item.specifications) as Specification[];
		return Array.isArray(specifications) ? specifications : [];
	} catch {
		return [];
	}
}

export const shippedFeatureCount = roadmap.items.filter(
	(item) => item.status === "Done",
).length;

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

export function relationshipsForItem(itemID: number) {
	return relationships.filter(
		(relationship) =>
			relationship.source_item_id === itemID ||
			relationship.target_item_id === itemID,
	);
}

export { roadmap };
