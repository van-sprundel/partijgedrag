import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function formatDate(date: Date): string {
	return new Intl.DateTimeFormat("nl-NL", {
		year: "numeric",
		month: "long",
		day: "numeric",
	}).format(date);
}

export function formatPercentage(value: number): string {
	return `${Math.round(value)}%`;
}

export const partyInfo: Record<string, { color: string; className: string }> = {
	VVD: { color: "#0066CC", className: "bg-blue-600" },
	D66: { color: "#00AA55", className: "bg-green-600" },
	PvdA: { color: "#DC143C", className: "bg-red-600" },
	SP: { color: "#FF0000", className: "bg-red-700" },
	PVV: { color: "#FFD700", className: "bg-yellow-500" },
	GL: { color: "#7CB518", className: "bg-green-500" },
	FVD: { color: "#8B0000", className: "bg-red-900" },
	CDA: { color: "#008B8B", className: "bg-teal-600" },
	CU: { color: "#0066FF", className: "bg-blue-500" },
	SGP: { color: "#FFA500", className: "bg-orange-500" },
};

export function getPartyColor(shortName: string): string {
	return partyInfo[shortName]?.color || "#6B7280";
}

export function getPartyColorClass(shortName: string): string {
	return partyInfo[shortName]?.className || "bg-gray-500";
}

export function shuffleArray<T>(array: T[]): T[] {
	const shuffled = [...array];
	for (let i = shuffled.length - 1; i > 0; i--) {
		const j = Math.floor(Math.random() * (i + 1));
		[shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
	}
	return shuffled;
}

export function sleep(ms: number): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

export function debounce<T extends (...args: any[]) => any>(
	func: T,
	wait: number,
): (...args: Parameters<T>) => void {
	let timeout: ReturnType<typeof setTimeout>;
	return (...args: Parameters<T>) => {
		clearTimeout(timeout);
		timeout = setTimeout(() => func(...args), wait);
	};
}

export function truncateText(text: string, maxLength: number): string {
	if (text.length <= maxLength) return text;
	return text.slice(0, maxLength).trim() + "...";
}

export function calculateProgress(current: number, total: number): number {
	return Math.round((current / total) * 100);
}
