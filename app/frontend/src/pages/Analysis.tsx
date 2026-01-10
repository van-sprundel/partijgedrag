import { useMemo, useState } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../components/ui/Card";
import {
	useCoalitionAlignment,
	useMPDeviations,
	useParties,
	useTopicTrends,
} from "../hooks/api";
import type { Party } from "../lib/api";
import { formatPercentage } from "../lib/utils";

type TabType = "coalition" | "deviations" | "topics";

const cabinetPresets = {
	"schoof-i": {
		label: "Schoof I",
		from: "2024-07-02",
		to: new Date().toISOString().split("T")[0],
	},
	"rutte-iv": { label: "Rutte IV", from: "2022-01-10", to: "2024-07-02" },
	"rutte-iii": { label: "Rutte III", from: "2017-10-26", to: "2022-01-10" },
	all: { label: "Alle tijd", from: "", to: "" },
};

export function AnalysisPage() {
	const [activeTab, setActiveTab] = useState<TabType>("coalition");
	const [cabinetPreset, setCabinetPreset] = useState<keyof typeof cabinetPresets>("schoof-i");

	const dateRange = useMemo(() => {
		const preset = cabinetPresets[cabinetPreset];
		if (!preset.from || !preset.to) return undefined;
		
		const parsedFrom = new Date(preset.from);
		const parsedTo = new Date(preset.to);

		if (Number.isNaN(parsedFrom.getTime()) || Number.isNaN(parsedTo.getTime())) {
			return undefined;
		}

		// To ensure the date range is inclusive of the end date, set the time to the end of the day.
		parsedTo.setUTCHours(23, 59, 59, 999);

		return {
			dateFrom: parsedFrom,
			dateTo: parsedTo,
		};
	}, [cabinetPreset]);

	const { data: parties } = useParties({ activeOnly: true });

	return (
		<div className="container mx-auto p-4">
			<h1 className="text-3xl font-bold mb-2">Analyse</h1>
			<p className="text-gray-600 mb-6">
				Inzichten in stemgedrag, coalitievorming en trends.
			</p>

			<Card className="mb-4">
				<CardHeader>
					<CardTitle>Periode</CardTitle>
				</CardHeader>
				<CardContent>
					<select
						value={cabinetPreset}
						onChange={(e) => setCabinetPreset(e.target.value as keyof typeof cabinetPresets)}
						className="input"
					>
						{Object.entries(cabinetPresets).map(([key, preset]) => (
							<option key={key} value={key}>
								{preset.label}
							</option>
						))}
					</select>
				</CardContent>
			</Card>

			<div className="flex border-b mb-4">
				<button
					type="button"
					className={`py-2 px-4 ${activeTab === "coalition" ? "border-b-2 border-blue-500 font-semibold" : "text-gray-500"}`}
					onClick={() => setActiveTab("coalition")}
				>
					Coalitie Analyse
				</button>
				<button
					type="button"
					className={`py-2 px-4 ${activeTab === "deviations" ? "border-b-2 border-blue-500 font-semibold" : "text-gray-500"}`}
					onClick={() => setActiveTab("deviations")}
				>
					Afwijkende Kamerleden
				</button>
				<button
					type="button"
					className={`py-2 px-4 ${activeTab === "topics" ? "border-b-2 border-blue-500 font-semibold" : "text-gray-500"}`}
					onClick={() => setActiveTab("topics")}
				>
					Onderwerp Trends
				</button>
			</div>

			{activeTab === "coalition" && (
				<CoalitionTab dateRange={dateRange} parties={parties || []} />
			)}
			{activeTab === "deviations" && <DeviationsTab dateRange={dateRange} />}
			{activeTab === "topics" && <TopicsTab dateRange={dateRange} />}
		</div>
	);
}

function CoalitionTab({
	dateRange,
	parties,
}: { dateRange?: { dateFrom: Date; dateTo: Date }; parties: Party[] }) {
	const { data: alignments, isLoading } = useCoalitionAlignment(dateRange);

	const matrix = useMemo(() => {
		const m: { [key: string]: { [key: string]: number | undefined } } = {};

		// Initialize matrix
		parties.forEach((p1) => {
			m[p1.shortName] = {};
			parties.forEach((p2) => {
				m[p1.shortName][p2.shortName] = p1.id === p2.id ? 100 : undefined;
			});
		});

		// Fill in alignment data
		alignments?.forEach((a) => {
			if (m[a.fractie1Name]) {
				m[a.fractie1Name][a.fractie2Name] = a.alignmentPct;
			}
			if (m[a.fractie2Name]) {
				m[a.fractie2Name][a.fractie1Name] = a.alignmentPct;
			}
		});

		return m;
	}, [parties, alignments]);

	const partyNames = parties.map((p) => p.shortName);

	const getCellColor = (value?: number) => {
		if (value === undefined) return "bg-gray-50";
		if (value > 80) return "bg-green-300";
		if (value > 60) return "bg-green-200";
		if (value >= 40) return "bg-gray-100";
		if (value < 20) return "bg-red-300";
		return "bg-red-200";
	};

	if (isLoading) {
		return (
			<Card>
				<CardContent className="p-8 text-center">Laden...</CardContent>
			</Card>
		);
	}

	if (!alignments || alignments.length === 0) {
		return (
			<Card>
				<CardContent className="p-8 text-center">
					Geen coalitie data beschikbaar. Voer eerst de ETL uit met{" "}
					<code className="bg-gray-100 px-2 py-1 rounded">analyze refresh</code>.
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle>Coalitie Alignment Matrix</CardTitle>
				<CardDescription>
					Percentage van de tijd dat partijen hetzelfde stemmen.
				</CardDescription>
			</CardHeader>
			<CardContent>
				<div className="overflow-x-auto">
					<table className="min-w-full border-collapse border text-sm">
						<thead>
							<tr>
								<th className="border p-2" />
								{partyNames.map((name) => (
									<th key={name} className="border p-2 text-xs">
										{name}
									</th>
								))}
							</tr>
						</thead>
						<tbody>
							{partyNames.map((p1) => (
								<tr key={p1}>
									<td className="border p-2 font-medium">{p1}</td>
									{partyNames.map((p2) => {
										const value = matrix[p1]?.[p2];
										return (
											<td
												key={p2}
												className={`border p-2 text-center ${p1 === p2 ? "bg-gray-200" : getCellColor(value)}`}
											>
												{p1 === p2
													? "-"
													: value === undefined
														? "-"
														: `${Math.round(value)}%`}
											</td>
										);
									})}
								</tr>
							))}
						</tbody>
					</table>
				</div>

				<div className="mt-6">
					<h3 className="font-semibold mb-2">Top 10 Meest Gealigneerde Paren</h3>
					<div className="grid gap-2">
						{alignments.slice(0, 10).map((a, i) => (
							<div
								key={`${a.fractie1Id}-${a.fractie2Id}`}
								className="flex items-center justify-between p-2 bg-gray-50 rounded"
							>
								<span>
									{i + 1}. {a.fractie1Name} & {a.fractie2Name}
								</span>
								<span className="font-bold text-green-600">
									{formatPercentage(a.alignmentPct)}
								</span>
							</div>
						))}
					</div>
				</div>
			</CardContent>
		</Card>
	);
}

function DeviationsTab({ dateRange }: { dateRange?: { dateFrom: Date; dateTo: Date } }) {
	const { data: deviations, isLoading } = useMPDeviations(dateRange, 25);

	if (isLoading) {
		return (
			<Card>
				<CardContent className="p-8 text-center">Laden...</CardContent>
			</Card>
		);
	}

	if (!deviations || deviations.length === 0) {
		return (
			<Card>
				<CardContent className="p-8 text-center">
					Geen afwijkingsdata beschikbaar.
				</CardContent>
			</Card>
		);
	}

	// All have 0% deviation - show that info
	const hasDeviations = deviations.some((d) => d.deviationPct > 0);

	return (
		<Card>
			<CardHeader>
				<CardTitle>Kamerleden die Afwijken van Partijlijn</CardTitle>
				<CardDescription>
					Kamerleden die het vaakst anders stemmen dan hun fractie.
				</CardDescription>
			</CardHeader>
			<CardContent>
				{!hasDeviations && (
					<div className="mb-4 p-4 bg-blue-50 rounded-lg text-blue-800">
						Alle kamerleden in deze periode stemmen 100% in lijn met hun fractie.
						Nederlandse fracties hebben sterke partijdiscipline.
					</div>
				)}
				<div className="overflow-x-auto">
					<table className="min-w-full border-collapse">
						<thead>
							<tr className="bg-gray-50">
								<th className="border p-3 text-left">Naam</th>
								<th className="border p-3 text-left">Partij</th>
								<th className="border p-3 text-right">Afwijking %</th>
								<th className="border p-3 text-right">Afwijkingen</th>
								<th className="border p-3 text-right">Totaal stemmen</th>
							</tr>
						</thead>
						<tbody>
							{deviations.map((d) => (
								<tr key={d.persoonId} className="hover:bg-gray-50">
									<td className="border p-3">{d.persoonNaam}</td>
									<td className="border p-3">
										<span className="px-2 py-1 bg-gray-100 rounded text-sm">
											{d.fractieNaam}
										</span>
									</td>
									<td className="border p-3 text-right">
										<span
											className={`font-medium ${d.deviationPct > 5 ? "text-red-600" : "text-green-600"}`}
										>
											{formatPercentage(d.deviationPct)}
										</span>
									</td>
									<td className="border p-3 text-right">{d.deviationCount}</td>
									<td className="border p-3 text-right">{d.totalVotes}</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</CardContent>
		</Card>
	);
}

function TopicsTab({ dateRange }: { dateRange?: { dateFrom: Date; dateTo: Date } }) {
	const { data: trends, isLoading } = useTopicTrends(dateRange);

	if (isLoading) {
		return (
			<Card>
				<CardContent className="p-8 text-center">Laden...</CardContent>
			</Card>
		);
	}

	if (!trends || trends.length === 0) {
		return (
			<Card>
				<CardContent className="p-8 text-center">
					Geen onderwerp trends beschikbaar.
				</CardContent>
			</Card>
		);
	}

	const maxCount = Math.max(...trends.map((t) => t.motionCount));

	return (
		<Card>
			<CardHeader>
				<CardTitle>Moties per Onderwerp</CardTitle>
				<CardDescription>
					Aantal moties per beleidsonderwerp, met aangenomen/verworpen
					uitsplitsing.
				</CardDescription>
			</CardHeader>
			<CardContent>
				<div className="space-y-3">
					{trends.map((t) => {
						const acceptedPct =
							t.motionCount > 0 ? (t.acceptedCount / t.motionCount) * 100 : 0;
						const rejectedPct =
							t.motionCount > 0 ? (t.rejectedCount / t.motionCount) * 100 : 0;
						const barWidth = (t.motionCount / maxCount) * 100;

						return (
							<div key={t.categoryId} className="space-y-1">
								<div className="flex justify-between items-center">
									<span className="font-medium">{t.categoryName}</span>
									<span className="text-sm text-gray-500">
										{t.motionCount} moties
									</span>
								</div>
								<div className="h-6 bg-gray-100 rounded overflow-hidden relative">
									<div
										className="h-full bg-gray-300 absolute left-0"
										style={{ width: `${barWidth}%` }}
									/>
									{t.acceptedCount > 0 && (
										<div
											className="h-full bg-green-500 absolute left-0"
											style={{
												width: `${(t.acceptedCount / maxCount) * 100}%`,
											}}
										/>
									)}
									{t.rejectedCount > 0 && (
										<div
											className="h-full bg-red-500 absolute"
											style={{
												left: `${(t.acceptedCount / maxCount) * 100}%`,
												width: `${(t.rejectedCount / maxCount) * 100}%`,
											}}
										/>
									)}
								</div>
								<div className="flex gap-4 text-xs text-gray-500">
									<span className="text-green-600">
										Aangenomen: {t.acceptedCount} ({Math.round(acceptedPct)}%)
									</span>
									<span className="text-red-600">
										Verworpen: {t.rejectedCount} ({Math.round(rejectedPct)}%)
									</span>
								</div>
							</div>
						);
					})}
				</div>
			</CardContent>
		</Card>
	);
}
