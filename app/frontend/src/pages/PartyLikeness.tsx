import { ArrowLeft, Users } from "lucide-react";
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
				</div>

				{similarities.length === 0 ? (
					<Card>
						<CardContent className="text-center py-12">
							<p className="text-gray-600">
								Niet genoeg stemdata beschikbaar om gelijkenissen te berekenen.
							</p>
						</CardContent>
					</Card>
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
