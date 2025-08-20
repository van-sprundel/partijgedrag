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

export function getPartyColor(shortName: string): string {
	const colors: Record<string, string> = {
		VVD: "#0066CC",
		D66: "#00AA55",
		PvdA: "#DC143C",
		SP: "#FF0000",
		PVV: "#FFD700",
		GL: "#7CB518",
		FVD: "#8B0000",
		CDA: "#008B8B",
		CU: "#0066FF",
		SGP: "#FFA500",
	};
	return colors[shortName] || "#6B7280";
}

export function getPartyColorClass(shortName: string): string {
	const classes: Record<string, string> = {
		VVD: "bg-blue-600",
		D66: "bg-green-600",
		PvdA: "bg-red-600",
		SP: "bg-red-700",
		PVV: "bg-yellow-500",
		GL: "bg-green-500",
		FVD: "bg-red-900",
		CDA: "bg-teal-600",
		CU: "bg-blue-500",
		SGP: "bg-orange-500",
	};
	return classes[shortName] || "bg-gray-500";
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

export function getVoteTypeLabel(voteType: string): string {
	const labels: Record<string, string> = {
		FOR: "Voor",
		AGAINST: "Tegen",
		ABSTAIN: "Onthouding",
		ABSENT: "Afwezig",
	};
	return labels[voteType] || voteType;
}

export function getVoteTypeColor(voteType: string): string {
	const colors: Record<string, string> = {
		FOR: "text-green-600",
		AGAINST: "text-red-600",
		ABSTAIN: "text-yellow-600",
		ABSENT: "text-gray-600",
	};
	return colors[voteType] || "text-gray-600";
}

export function generateSessionId(): string {
	return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}
