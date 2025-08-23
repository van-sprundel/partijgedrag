import { ArrowLeft, Grid, List, Users } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
import { Button } from "../components/ui/Button";
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from "../components/ui/Card";
import { usePartySimilarity } from "../hooks/api";
import type { PartySimilarity } from "../lib/api";

export function PartyLikenessPage() {
	const { data, isLoading, error } = usePartySimilarity(true, 10);
	const [viewMode, setViewMode] = useState<"list" | "matrix">("matrix");

	if (isLoading) {
		return (
			<div className="min-h-screen bg-gradient-to-b from-primary-50 to-white">
				<div className="container mx-auto px-4 py-16">
					<div className="text-center">
						<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
						<p className="text-gray-600">
							Bezig met analyseren van stemgedrag...
						</p>
					</div>
				</div>
			</div>
		);
	}

	if (error || !data) {
		return (
			<div className="min-h-screen bg-gradient-to-b from-primary-50 to-white">
				<div className="container mx-auto px-4 py-16">
					<div className="text-center">
						<h1 className="text-2xl font-bold text-gray-900 mb-4">
							Er ging iets mis
						</h1>
						<p className="text-gray-600 mb-8">
							We konden de partijgelijkenissen niet laden.
						</p>
						<Link to="/">
							<Button>Terug naar home</Button>
						</Link>
					</div>
				</div>
			</div>
		);
	}

	const { similarities, totalMotions } = data;

	return (
		<div className="min-h-screen bg-gradient-to-b from-primary-50 to-white">
			<div className="container mx-auto px-4 py-8">
				{/* Header */}
				<div className="mb-8">
					<Link
						to="/"
						className="inline-flex items-center text-primary-600 hover:text-primary-700 mb-4"
					>
						<ArrowLeft className="h-4 w-4 mr-2" />
						Terug naar home
					</Link>

					<div className="flex items-center mb-4">
						<div className="p-3 bg-primary-100 rounded-full mr-4">
							<Users className="h-8 w-8 text-primary-600" />
						</div>
						<div>
							<h1 className="text-4xl font-bold text-gray-900">
								Partijgelijkenissen
							</h1>
							<p className="text-xl text-gray-600">
								Ontdek welke partijen het meest overeenkomen in hun stemgedrag
							</p>
						</div>
					</div>

					<div className="bg-blue-50 border-l-4 border-blue-400 p-4 rounded">
						<p className="text-blue-800 text-sm">
							<strong>Methode:</strong> De gelijkenis wordt berekend op basis
							van {totalMotions} moties waarin partijen daadwerkelijk hebben
							gestemd. Een score van 100% betekent dat twee partijen altijd
							hetzelfde hebben gestemd.
						</p>
					</div>

					{/* View mode toggle */}
					<div className="flex justify-center mb-8">
						<div className="bg-white p-1 rounded-lg shadow-sm border">
							<button
								onClick={() => setViewMode("matrix")}
								className={`inline-flex items-center px-4 py-2 rounded-md text-sm font-medium transition-colors ${
									viewMode === "matrix"
										? "bg-primary-600 text-white"
										: "text-gray-700 hover:text-primary-600"
								}`}
							>
								<Grid className="h-4 w-4 mr-2" />
								Matrix weergave
							</button>
							<button
								onClick={() => setViewMode("list")}
								className={`inline-flex items-center px-4 py-2 rounded-md text-sm font-medium transition-colors ${
									viewMode === "list"
										? "bg-primary-600 text-white"
										: "text-gray-700 hover:text-primary-600"
								}`}
							>
								<List className="h-4 w-4 mr-2" />
								Lijst weergave
							</button>
						</div>
					</div>
				</div>

				{similarities.length === 0 ? (
					<Card>
						<CardContent className="text-center py-12">
							<p className="text-gray-600">
								Niet genoeg stemdata beschikbaar om gelijkenissen te berekenen.
							</p>
						</CardContent>
					</Card>
				) : viewMode === "matrix" ? (
					<PartyMatrix similarities={similarities} />
				) : (
					<>
						{/* Top matches */}
						<div className="mb-8">
							<h2 className="text-2xl font-bold mb-4 text-gray-900">
								Meest vergelijkbare partijen
							</h2>
							<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
								{similarities
									.slice(0, 6)
									.map((sim: PartySimilarity, index: number) => (
										<SimilarityCard
											key={`${sim.party1.id}-${sim.party2.id}`}
											similarity={sim}
											rank={index + 1}
										/>
									))}
							</div>
						</div>

						{/* Full similarity table */}
						<Card>
							<CardHeader>
								<CardTitle>Alle partijvergelijkingen</CardTitle>
							</CardHeader>
							<CardContent>
								<div className="overflow-x-auto">
									<table className="w-full text-sm">
										<thead>
											<tr className="border-b">
												<th className="text-left py-3 px-4">Rang</th>
												<th className="text-left py-3 px-4">Partij 1</th>
												<th className="text-left py-3 px-4">Partij 2</th>
												<th className="text-left py-3 px-4">Gelijkenis</th>
												<th className="text-left py-3 px-4">Overeenstemming</th>
											</tr>
										</thead>
										<tbody>
											{similarities.map(
												(sim: PartySimilarity, index: number) => (
													<tr
														key={`${sim.party1.id}-${sim.party2.id}-${index}`}
														className="border-b hover:bg-gray-50"
													>
														<td className="py-3 px-4 font-medium">
															#{index + 1}
														</td>
														<td className="py-3 px-4">
															<div className="flex items-center">
																<div
																	className="w-3 h-3 rounded-full mr-2"
																	style={{
																		backgroundColor:
																			sim.party1.color || "#6B7280",
																	}}
																></div>
																{sim.party1.shortName}
															</div>
														</td>
														<td className="py-3 px-4">
															<div className="flex items-center">
																<div
																	className="w-3 h-3 rounded-full mr-2"
																	style={{
																		backgroundColor:
																			sim.party2.color || "#6B7280",
																	}}
																></div>
																{sim.party2.shortName}
															</div>
														</td>
														<td className="py-3 px-4">
															<div className="flex items-center">
																<div className="flex-1 bg-gray-200 rounded-full h-2 mr-3">
																	<div
																		className="bg-gradient-to-r from-red-500 via-yellow-500 to-green-500 h-2 rounded-full"
																		style={{ width: `${sim.similarity}%` }}
																	></div>
																</div>
																<span className="font-medium text-sm">
																	{sim.similarity}%
																</span>
															</div>
														</td>
														<td className="py-3 px-4 text-gray-600">
															{sim.agreementCount} van {sim.totalComparisons}{" "}
															moties
														</td>
													</tr>
												),
											)}
										</tbody>
									</table>
								</div>
							</CardContent>
						</Card>
					</>
				)}

				{/* Info section */}
				<div className="mt-12 bg-white rounded-2xl p-8 shadow-sm">
					<h2 className="text-2xl font-bold mb-6 text-gray-900">
						Hoe wordt gelijkenis berekend?
					</h2>
					<div className="grid md:grid-cols-2 gap-8">
						<div>
							<h3 className="font-semibold mb-3 text-lg">Methodologie</h3>
							<ul className="space-y-2 text-gray-600">
								<li>
									• We vergelijken hoe partijen hebben gestemd op dezelfde
									moties
								</li>
								<li>
									• Alleen moties met stemmen van beide partijen worden
									meegenomen
								</li>
								<li>
									• Een partij's positie wordt bepaald door de meerderheid van
									hun leden
								</li>
								<li>
									• De score is het percentage waarbij partijen het eens waren
								</li>
							</ul>
						</div>
						<div>
							<h3 className="font-semibold mb-3 text-lg">Interpretatie</h3>
							<ul className="space-y-2 text-gray-600">
								<li>
									• <span className="font-medium">90-100%:</span> Zeer
									vergelijkbaar stemgedrag
								</li>
								<li>
									• <span className="font-medium">70-89%:</span> Grotendeels
									vergelijkbaar
								</li>
								<li>
									• <span className="font-medium">50-69%:</span> Gedeeltelijk
									vergelijkbaar
								</li>
								<li>
									• <span className="font-medium">0-49%:</span> Vaak
									verschillend stemgedrag
								</li>
							</ul>
						</div>
					</div>
				</div>

				{/* CTA back to compass */}
				<div className="text-center mt-12">
					<h2 className="text-2xl font-bold mb-4 text-gray-900">
						Vind jouw persoonlijke match
					</h2>
					<p className="text-gray-600 mb-6">
						Ontdek welke partij het beste aansluit bij jouw standpunten
					</p>
					<Link to="/compass">
						<Button size="lg">Start de stemwijzer</Button>
					</Link>
				</div>
			</div>
		</div>
	);
}

function SimilarityCard({
	similarity,
	rank,
}: {
	similarity: PartySimilarity;
	rank: number;
}) {
	const getSimilarityColor = (score: number) => {
		if (score >= 90) return "bg-green-500";
		if (score >= 70) return "bg-yellow-500";
		if (score >= 50) return "bg-orange-500";
		return "bg-red-500";
	};

	const getSimilarityDescription = (score: number) => {
		if (score >= 90) return "Zeer vergelijkbaar";
		if (score >= 70) return "Grotendeels vergelijkbaar";
		if (score >= 50) return "Gedeeltelijk vergelijkbaar";
		return "Vaak verschillend";
	};

	return (
		<Card className="hover:shadow-lg transition-shadow">
			<CardContent className="p-6">
				<div className="flex items-center justify-between mb-4">
					<div className="flex items-center space-x-2">
						<div className="w-6 h-6 bg-primary-100 rounded-full flex items-center justify-center">
							<span className="text-xs font-bold text-primary-600">
								#{rank}
							</span>
						</div>
					</div>
					<div className="text-right">
						<div className="text-2xl font-bold text-gray-900">
							{similarity.similarity}%
						</div>
						<div className="text-xs text-gray-500">
							{getSimilarityDescription(similarity.similarity)}
						</div>
					</div>
				</div>

				<div className="space-y-3">
					<div className="flex items-center">
						<div
							className="w-3 h-3 rounded-full mr-2"
							style={{ backgroundColor: similarity.party1.color || "#6B7280" }}
						></div>
						<span className="font-medium">{similarity.party1.shortName}</span>
					</div>
					<div className="text-center text-gray-400 text-sm">vs</div>
					<div className="flex items-center">
						<div
							className="w-3 h-3 rounded-full mr-2"
							style={{ backgroundColor: similarity.party2.color || "#6B7280" }}
						></div>
						<span className="font-medium">{similarity.party2.shortName}</span>
					</div>
				</div>

				<div className="mt-4">
					<div className="flex items-center justify-between text-xs text-gray-500 mb-1">
						<span>Overeenstemming</span>
						<span>
							{similarity.agreementCount}/{similarity.totalComparisons}
						</span>
					</div>
					<div className="w-full bg-gray-200 rounded-full h-2">
						<div
							className={`h-2 rounded-full ${getSimilarityColor(similarity.similarity)}`}
							style={{ width: `${similarity.similarity}%` }}
						></div>
					</div>
				</div>
			</CardContent>
		</Card>
	);
}

function PartyMatrix({ similarities }: { similarities: PartySimilarity[] }) {
	// Get all unique parties
	const parties = Array.from(
		new Set(
			[
				...similarities.map((s) => s.party1),
				...similarities.map((s) => s.party2),
			].map((p) => p.id),
		),
	)
		.map((id) => {
			const party =
				similarities.find((s) => s.party1.id === id)?.party1 ||
				similarities.find((s) => s.party2.id === id)?.party2;
			return party!;
		})
		.sort((a, b) => a.shortName.localeCompare(b.shortName));

	// Create similarity lookup map
	const similarityMap = new Map<string, PartySimilarity>();
	similarities.forEach((sim) => {
		const key1 = `${sim.party1.id}-${sim.party2.id}`;
		const key2 = `${sim.party2.id}-${sim.party1.id}`;
		similarityMap.set(key1, sim);
		similarityMap.set(key2, sim);
	});

	const getSimilarity = (party1Id: string, party2Id: string): number | null => {
		if (party1Id === party2Id) return 100; // Same party = 100% similarity
		const sim = similarityMap.get(`${party1Id}-${party2Id}`);
		return sim ? sim.similarity : null;
	};

	const getSimilarityColor = (score: number | null) => {
		if (score === null) return "bg-gray-100";
		if (score === 100) return "bg-gray-800"; // Same party
		if (score >= 90) return "bg-green-500";
		if (score >= 80) return "bg-green-400";
		if (score >= 70) return "bg-yellow-400";
		if (score >= 60) return "bg-orange-400";
		if (score >= 50) return "bg-orange-500";
		return "bg-red-500";
	};

	const getSimilarityTextColor = (score: number | null) => {
		if (score === null) return "text-gray-400";
		if (score === 100) return "text-white";
		if (score >= 70) return "text-white";
		return "text-white";
	};

	return (
		<Card>
			<CardHeader>
				<CardTitle>Partijgelijkenis Matrix</CardTitle>
				<p className="text-sm text-gray-600">
					Vergelijk alle partijen onderling. Klik op een cel voor details.
				</p>
			</CardHeader>
			<CardContent>
				<div className="overflow-x-auto">
					<div className="min-w-full">
						{/* Header row */}
						<div className="flex">
							<div className="w-24 h-12 flex-shrink-0"></div>
							{parties.map((party) => (
								<div
									key={party.id}
									className="w-16 h-12 flex-shrink-0 flex items-center justify-center border-b border-gray-200"
								>
									<div className="transform -rotate-45 text-xs font-medium text-center whitespace-nowrap">
										{party.shortName}
									</div>
								</div>
							))}
						</div>

						{/* Matrix rows */}
						{parties.map((rowParty) => (
							<div key={rowParty.id} className="flex">
								{/* Row header */}
								<div className="w-24 h-12 flex-shrink-0 flex items-center px-2 border-r border-gray-200">
									<div className="flex items-center">
										<div
											className="w-3 h-3 rounded-full mr-2 flex-shrink-0"
											style={{ backgroundColor: rowParty.color || "#6B7280" }}
										></div>
										<span className="text-xs font-medium truncate">
											{rowParty.shortName}
										</span>
									</div>
								</div>

								{/* Matrix cells */}
								{parties.map((colParty) => {
									const similarity = getSimilarity(rowParty.id, colParty.id);
									const simData = similarityMap.get(
										`${rowParty.id}-${colParty.id}`,
									);

									return (
										<div
											key={colParty.id}
											className={`w-16 h-12 flex-shrink-0 border border-gray-200 flex items-center justify-center cursor-pointer hover:border-gray-400 transition-colors ${getSimilarityColor(similarity)}`}
											title={
												similarity !== null
													? `${rowParty.shortName} vs ${colParty.shortName}: ${similarity}%${
															simData
																? ` (${simData.agreementCount}/${simData.totalComparisons} moties)`
																: ""
														}`
													: "Geen data beschikbaar"
											}
										>
											<span
												className={`text-xs font-bold ${getSimilarityTextColor(similarity)}`}
											>
												{similarity !== null ? `${similarity}%` : "-"}
											</span>
										</div>
									);
								})}
							</div>
						))}
					</div>
				</div>

				{/* Legend */}
				<div className="mt-6 p-4 bg-gray-50 rounded-lg">
					<h4 className="font-medium mb-3">Legenda</h4>
					<div className="flex flex-wrap gap-4 text-sm">
						<div className="flex items-center">
							<div className="w-4 h-4 bg-green-500 rounded mr-2"></div>
							<span>90-100%: Zeer vergelijkbaar</span>
						</div>
						<div className="flex items-center">
							<div className="w-4 h-4 bg-yellow-400 rounded mr-2"></div>
							<span>70-89%: Grotendeels vergelijkbaar</span>
						</div>
						<div className="flex items-center">
							<div className="w-4 h-4 bg-orange-500 rounded mr-2"></div>
							<span>50-69%: Gedeeltelijk vergelijkbaar</span>
						</div>
						<div className="flex items-center">
							<div className="w-4 h-4 bg-red-500 rounded mr-2"></div>
							<span>0-49%: Vaak verschillend</span>
						</div>
						<div className="flex items-center">
							<div className="w-4 h-4 bg-gray-800 rounded mr-2"></div>
							<span>Zelfde partij</span>
						</div>
					</div>
				</div>
			</CardContent>
		</Card>
	);
}
