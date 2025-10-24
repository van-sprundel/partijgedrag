const SESSION_STORAGE_KEY = "partijgedrag_sessions";
const MAX_SESSIONS = 10;

export interface StoredSession {
	id: string;
	createdAt: Date;
}

export function saveSessionId(sessionId: string): void {
	try {
		const sessions = getStoredSessions();
		const newSession: StoredSession = {
			id: sessionId,
			createdAt: new Date(),
		};

		// Remove duplicate if exists
		const filtered = sessions.filter((s) => s.id !== sessionId);

		// Add new session at the beginning
		const updated = [newSession, ...filtered].slice(0, MAX_SESSIONS);

		localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(updated));
	} catch (error) {
		console.error("Failed to save session ID:", error);
	}
}

export function getStoredSessions(): StoredSession[] {
	try {
		const stored = localStorage.getItem(SESSION_STORAGE_KEY);
		if (!stored) return [];

		const parsed = JSON.parse(stored);
		// Convert date strings back to Date objects
		return parsed.map((s: { id: string; createdAt: string }) => ({
			id: s.id,
			createdAt: new Date(s.createdAt),
		}));
	} catch (error) {
		console.error("Failed to get stored sessions:", error);
		return [];
	}
}

export function removeSessionId(sessionId: string): void {
	try {
		const sessions = getStoredSessions();
		const filtered = sessions.filter((s) => s.id !== sessionId);
		localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(filtered));
	} catch (error) {
		console.error("Failed to remove session ID:", error);
	}
}

export function clearAllSessions(): void {
	try {
		localStorage.removeItem(SESSION_STORAGE_KEY);
	} catch (error) {
		console.error("Failed to clear sessions:", error);
	}
}
