import { useId, useMemo, useState } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../components/ui/Card";
import {
	useMotionCategories,
	useParties,
	usePartyCategoryLikeness,
	usePartyFocus,
	usePartyLikenessMatrix,
} from "../hooks/api";
import type { MotionCategory, Party } from "../lib/api";
import { formatPercentage, getPartyColorClass } from "../lib/utils";

const cabinetPresets = {
	"schoof-i": {
		from: "2024-07-02",
		to: new Date().toISOString().split("T")[0],
	},
	"rutte-iv": { from: "2022-01-10", to: "2024-07-02" },
	"rutte-iii": { from: "2017-10-26", to: "2022-01-10" },
	"rutte-asscher": { from: "2012-11-05", to: "2017-10-26" },
	"rutte-verhagen": { from: "2010-10-14", to: "2012-11-05" },
	"balkenende-iv": { from: "2007-02-22", to: "2010-10-14" },
	"balkenende-iii": { from: "2006-07-07", to: "2007-02-22" },
};

export function PartyLikenessPage() {
	const [activeTab, setActiveTab] = useState("matrix");
	const [dateFrom, setDateFrom] = useState<string>("");
	const [dateTo, setDateTo] = useState<string>("");

	const { data: parties, isLoading: isLoadingParties } = useParties();

	const handleTabClick = (partyId: string) => {
		setActiveTab(partyId);
	};

	const handlePresetChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
		if (!e.target.value) {
			setDateFrom("");
			setDateTo("");
			return;
		}
		const preset =
			cabinetPresets[e.target.value as keyof typeof cabinetPresets];
		if (preset) {
			setDateFrom(preset.from);
			setDateTo(preset.to);
		}
	};

	const filters =
		dateFrom && dateTo
			? { dateFrom: new Date(dateFrom), dateTo: new Date(dateTo) }
			: {};

	return (
		<div className="container mx-auto p-4">
			<Card className="mb-4">
				<CardHeader>
					<CardTitle>Filters</CardTitle>
					<CardDescription>
						Selecteer een periode om de analyses te filteren.
					</CardDescription>
				</CardHeader>
				<CardContent className="flex items-center gap-4">
					<div>
						<label
							htmlFor="preset"
							className="block text-sm font-medium text-gray-700"
						>
							Presets
						</label>
						<select
							id={useId()}
							onChange={handlePresetChange}
							className="input"
						>
							<option value="">Alles</option>
							<option value="schoof-i">Schoof I</option>
							<option value="rutte-iv">Rutte IV</option>
							<option value="rutte-iii">Rutte III</option>
							<option value="rutte-ascher">Rutte-Asscher</option>
							<option value="rutte-verhagen">Rutte-Verhagen</option>
							<option value="balkenende-iv">Balkenende IV</option>
							<option value="balkenende-iii">Balkenende III</option>
						</select>
					</div>
					<div>
						<label
							htmlFor="dateFrom"
							className="block text-sm font-medium text-gray-700"
						>
							Vanaf
						</label>
						<input
							type="date"
							id={useId()}
							value={dateFrom}
							onChange={(e) => setDateFrom(e.target.value)}
							className="input"
						/>
					</div>
					<div>
						<label
							htmlFor="dateTo"
							className="block text-sm font-medium text-gray-700"
						>
							Tot
						</label>
						<input
							type="date"
							id={useId()}
							value={dateTo}
							onChange={(e) => setDateTo(e.target.value)}
							className="input"
						/>
					</div>
				</CardContent>
			</Card>

			<h1 className="text-3xl font-bold mb-4">Partijgelijkenis</h1>
			<div className="flex border-b mb-4 overflow-x-auto">
				<button
					type="button"
					className={`py-2 px-4 whitespace-nowrap ${activeTab === "matrix" ? "border-b-2 border-blue-500 font-semibold" : "text-gray-500"}`}
					onClick={() => handleTabClick("matrix")}
				>
					Gelijkenis Matrix
				</button>
				{parties?.map((party) => (
					<button
						type="button"
						key={party.id}
						className={`py-2 px-4 whitespace-nowrap ${activeTab === party.id ? "border-b-2 border-blue-500 font-semibold" : "text-gray-500"}`}
						onClick={() => handleTabClick(party.id)}
					>
						{party.shortName}
					</button>
				))}
			</div>

			<div>
				{isLoadingParties ? (
					<div>Loading...</div>
				) : (
					<>
						{activeTab === "matrix" && (
							<LikenessMatrix parties={parties || []} filters={filters} />
						)}
						{activeTab !== "matrix" && (
							<PartyFocusTab
								partyId={activeTab}
								parties={parties || []}
								filters={filters}
							/>
						)}
					</>
				)}
			</div>
		</div>
	);
}

function LikenessMatrix({
	parties,
	filters,
}: {
	parties: Party[];
	filters: any;
}) {
	const { data: likenessData, isLoading } = usePartyLikenessMatrix(filters);

	const matrix = useMemo(() => {
		const m: { [key: string]: { [key: string]: number } } = {};
		parties.forEach((p1) => {
			m[p1.id] = {};
			parties.forEach((p2) => {
				if (p1.id === p2.id) m[p1.id][p2.id] = 100;
				else m[p1.id][p2.id] = 0;
			});
		});
		likenessData?.forEach((entry) => {
			if (m[entry.party1Id])
				m[entry.party1Id][entry.party2Id] = entry.likenessPercentage;
			if (m[entry.party2Id])
				m[entry.party2Id][entry.party1Id] = entry.likenessPercentage;
		});
		return m;
	}, [parties, likenessData]);

	const getCellColor = (value: number) => {
		if (value > 80) return "bg-green-300";
		if (value > 60) return "bg-green-200";
		if (value < 40) return "bg-red-200";
		if (value < 20) return "bg-red-300";
		return "bg-white";
	};

	if (isLoading) {
		return <div>Loading matrix...</div>;
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle>Partijgelijkenis Matrix</CardTitle>
				<CardDescription>Hoe vaak stemmen partijen hetzelfde?</CardDescription>
			</CardHeader>
			<CardContent>
				<div className="overflow-x-auto">
					<table className="min-w-full border-collapse border">
						<thead>
							<tr>
								<th className="border p-2"></th>
								{parties.map((party) => (
									<th key={party.id} className="border p-2">
										<div className="flex items-center justify-center">
											<img
												src={`data:${party.contentType};base64,${party.logoData}`}
												alt={` `}
												className="w-6 h-6 object-contain"
											/>
										</div>
									</th>
								))}
							</tr>
						</thead>
						<tbody>
							{parties.map((p1) => (
								<tr key={p1.id}>
									<td className="border p-2 font-bold">
										<div className="flex items-center gap-2">
											<img
												src={`data:${p1.contentType};base64,${p1.logoData}`}
												alt={` `}
												className="w-6 h-6 object-contain"
											/>
											<span>{p1.name}</span>
										</div>
									</td>
									{parties.map((p2) => (
										<td
											key={p2.id}
											className={`border p-2 text-center ${p1.id === p2.id ? "bg-gray-100" : getCellColor(matrix[p1.id]?.[p2.id] ?? 0)}`}
										>
											{p1.id === p2.id
												? "-"
												: formatPercentage(matrix[p1.id]?.[p2.id] ?? 0)}
										</td>
									))}
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</CardContent>
		</Card>
	);
}

function PartyFocusTab({
	partyId,
	parties,
	filters,
}: {
	partyId: string;
	parties: Party[];
	filters: any;
}) {
	const { data: focusData, isLoading: isLoadingFocus } = usePartyFocus(
		partyId,
		filters,
	);

	if (isLoadingFocus) {
		return <div>Loading party focus...</div>;
	}

	if (!focusData) {
		return <div>No data for this party.</div>;
	}

	const { party, categories } = focusData;

	return (
		<div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
			<div className="lg:col-span-1">
				<Card>
					<CardHeader>
						<CardTitle>Partijfocus voor {party.name}</CardTitle>
						<CardDescription>
							Aantal moties ingediend per categorie.
						</CardDescription>
					</CardHeader>
					<CardContent>
						{categories.length > 0 ? (
							<ul className="space-y-2">
								{categories.map((cat) => (
									<li
										key={cat.categoryId}
										className="flex justify-between items-center p-2 bg-gray-50 rounded-lg"
									>
										<span className="font-medium">{cat.categoryName}</span>
										<span className="font-bold">{cat.motionCount}</span>
									</li>
								))}
							</ul>
						) : (
							<p>
								Deze partij heeft geen moties ingediend in de geselecteerde
								periode of er zijn geen gecategoriseerde moties.
							</p>
						)}
					</CardContent>
				</Card>
			</div>
			<div className="lg:col-span-2">
				<PartyCategoryLikenessMatrix
					partyId={partyId}
					parties={parties}
					filters={filters}
				/>
			</div>
		</div>
	);
}

function PartyCategoryLikenessMatrix({
	partyId,
	parties,
	filters,
}: {
	partyId: string;
	parties: Party[];
	filters: any;
}) {
	const { data: categoryLikenessData, isLoading: isLoadingCategoryLikeness } =
		usePartyCategoryLikeness(partyId, filters);
	const { data: categories, isLoading: isLoadingCategories } =
		useMotionCategories();

	const otherParties = useMemo(
		() => parties.filter((p) => p.id !== partyId),
		[parties, partyId],
	);

	const matrix = useMemo(() => {
		const m: { [key: string]: { [key: string]: number } } = {};
		categories?.forEach((cat) => {
			m[cat.id] = {};
			otherParties.forEach((p) => {
				m[cat.id][p.id] = 0;
			});
		});
		categoryLikenessData?.forEach((entry) => {
			if (m[entry.categoryId]) {
				m[entry.categoryId][entry.party2Id] = entry.likenessPercentage;
			}
		});
		return m;
	}, [categories, otherParties, categoryLikenessData]);

	if (isLoadingCategoryLikeness || isLoadingCategories) {
		return <div>Loading category likeness matrix...</div>;
	}

	const getCellColor = (value: number) => {
		if (value > 80) return "bg-green-300";
		if (value > 60) return "bg-green-200";
		if (value >= 40) return "bg-gray-100";
		if (value < 20) return "bg-red-300";
		if (value < 40) return "bg-red-200";
		return "bg-white";
	};

	return (
		<Card>
			<CardHeader>
				<CardTitle>Gelijkenis per Categorie</CardTitle>
				<CardDescription>
					Hoe vaak stemt deze partij hetzelfde als andere partijen per
					onderwerp?
				</CardDescription>
			</CardHeader>
			<CardContent>
				<div className="overflow-x-auto">
					<table className="min-w-full border-collapse border">
						<thead>
							<tr>
								<th className="border p-2 text-left">Categorie</th>
								{otherParties.map((party) => (
									<th key={party.id} className="border p-2">
										<div className="flex items-center justify-center">
											<img
												src={`data:${party.contentType};base64,${party.logoData}`}
												alt={` `}
												className="w-6 h-6 object-contain"
											/>
										</div>
									</th>
								))}
							</tr>
						</thead>
						<tbody>
							{categories?.map((cat) => (
								<tr key={cat.id}>
									<td className="border p-2 font-medium">{cat.name}</td>
									{otherParties.map((p2) => {
										const likeness = matrix[cat.id]?.[p2.id] ?? 0;
										return (
											<td
												key={p2.id}
												className={`border p-2 text-center ${getCellColor(likeness)}`}
											>
												{formatPercentage(likeness)}
											</td>
										);
									})}
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</CardContent>
		</Card>
	);
}
