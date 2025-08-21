export function mapZaakToMotion(zaak: any, dossier?: any) {
	return {
		id: zaak.id,
		title: zaak.titel || zaak.onderwerp || "Untitled Motion",
		description: zaak.onderwerp,
		shortTitle: zaak.citeertitel,
		motionNumber: zaak.nummer,
		date: zaak.datum,
		status: zaak.status || "unknown",
		category: zaak.soort,
		bulletPoints: dossier?.bullet_points
			? Array.isArray(dossier.bullet_points)
				? dossier.bullet_points
				: []
			: [],
		originalId: zaak.id,
		createdAt: zaak.gestart_op || new Date(),
		updatedAt: zaak.gewijzigd_op || new Date(),
	};
}
