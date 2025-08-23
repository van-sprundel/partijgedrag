import {
	ArrowLeft,
	ArrowRight,
	ChevronDown,
	ChevronUp,
	RotateCcw,
	Share2,
} from "lucide-react";
import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Button } from "../components/ui/Button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../components/ui/Card";
import { Progress } from "../components/ui/Progress";
import { useCompassResults } from "../hooks/api";
import type { MotionDetail, PartyResult, VoteType } from "../lib/api";
import { formatPercentage, getPartyColorClass } from "../lib/utils";

export function ResultsPage() {
	const { sessionId } = useParams<{ sessionId: string }>();
	const [showAllParties, setShowAllParties] = useState(false);
	const [expandedMotions, setExpandedMotions] = useState<Set<string>>(
		new Set(),
	);
	const {
		data: results,
		isLoading,
		error,
	} = useCompassResults(sessionId || "");

	if (isLoading) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<div className="text-center">
					<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
					<p className="text-gray-600">Resultaten berekenen...</p>
				</div>
			</div>
		);
	}

	if (error || !results) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<Card className="max-w-md mx-auto">
					<CardHeader>
						<CardTitle>Resultaten niet gevonden</CardTitle>
						<CardDescription>
							De resultaten konden niet worden geladen. Mogelijk is de sessie
							verlopen.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="flex gap-2">
							<Link to="/compass">
								<Button variant="primary">Opnieuw beginnen</Button>
							</Link>
							<Link to="/">
								<Button variant="secondary">Terug naar home</Button>
							</Link>
						</div>
					</CardContent>
				</Card>
			</div>
		);
	}

	const topResults = results.partyResults.slice(0, 3);
	const otherResults = results.partyResults.slice(3);
	const displayedResults = showAllParties ? results.partyResults : topResults;

	const handleShare = async () => {
		if (navigator.share) {
			try {
				await navigator.share({
					title: "Mijn Partijgedrag resultaten",
					text: `Ik heb ${formatPercentage(topResults[0]?.agreement || 0)} overeenkomst met ${topResults[0]?.party.shortName}!`,
					url: window.location.href,
				});
			} catch (err) {
				console.log("Sharing failed:", err);
			}
		} else {
			// Fallback: copy to clipboard
			await navigator.clipboard.writeText(window.location.href);
			// You could show a toast notification here
		}
	};

	const toggleMotionDetails = (motionId: string) => {
		setExpandedMotions((prev) => {
			const newSet = new Set(prev);
			if (newSet.has(motionId)) {
				newSet.delete(motionId);
			} else {
				newSet.add(motionId);
			}
			return newSet;
		});
	};

	const getUserAnswerEmoji = (answer: string) => {
		switch (answer) {
			case "agree":
				return "👍";
			case "disagree":
				return "👎";
			case "neutral":
				return "🤷";
			default:
				return "❓";
		}
	};

	const getVoteEmoji = (position: VoteType) => {
		switch (position) {
			case "FOR":
				return "✅";
			case "AGAINST":
				return "❌";
			case "NEUTRAL":
				return "⚪";
			default:
				return "❓";
		}
	};

	return (
		<div className="min-h-screen bg-gray-50">
			{/* Header */}
			<div className="bg-white shadow-sm">
				<div className="container mx-auto px-4 py-4">
					<div className="flex items-center justify-between">
						<Link
							to="/"
							className="flex items-center text-gray-600 hover:text-gray-900"
						>
							<ArrowLeft className="h-5 w-5 mr-2" />
							Terug naar home
						</Link>
						<div className="flex items-center gap-2">
							<Button
								variant="ghost"
								size="sm"
								onClick={handleShare}
								className="flex items-center"
							>
								<Share2 className="h-4 w-4 mr-1" />
								Delen
							</Button>
							<Link to="/compass">
								<Button variant="ghost" size="sm" className="flex items-center">
									<RotateCcw className="h-4 w-4 mr-1" />
									Opnieuw
								</Button>
							</Link>
						</div>
					</div>
				</div>
			</div>

			<div className="container mx-auto px-4 py-8">
				{/* Results Header */}
				<div className="text-center mb-8">
					<h1 className="text-4xl font-bold text-gray-900 mb-4">
						Jouw resultaten
					</h1>
					<p className="text-xl text-gray-600">
						Gebaseerd op {results.totalAnswers} beantwoorde stellingen
					</p>
				</div>

				{/* Top Match */}
				{topResults[0] && (
					<Card className="mb-8 border-2 border-primary-200 bg-gradient-to-r from-primary-50 to-white">
						<CardHeader>
							<div className="flex items-center justify-between">
								<div>
									<CardTitle className="text-2xl mb-2">
										🏆 Beste match
									</CardTitle>
									<CardDescription className="text-lg">
										{topResults[0].party.name}
									</CardDescription>
								</div>
								<div className="text-right">
									<div className="text-4xl font-bold text-primary-600 mb-1">
										{formatPercentage(topResults[0].agreement)}
									</div>
									<div className="text-sm text-gray-600">overeenkomst</div>
								</div>
							</div>
						</CardHeader>
						<CardContent>
							<div className="flex items-center justify-between text-sm text-gray-600 mb-4">
								<span>
									{topResults[0].matchingVotes} van {topResults[0].totalVotes}{" "}
									stellingen overeenkomen
								</span>
								<span className="font-medium">
									Score: {topResults[0].score.toFixed(1)}
								</span>
							</div>
							<Progress
								value={topResults[0].agreement}
								variant="success"
								size="md"
							/>
						</CardContent>
					</Card>
				)}

				{/* All Results */}
				<Card className="mb-8">
					<CardHeader>
						<CardTitle>Alle partijen</CardTitle>
						<CardDescription>
							Hoe goed sluiten de verschillende partijen aan bij jouw
							standpunten?
						</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="space-y-4">
							{displayedResults.map((result: PartyResult, index: number) => (
								<div
									key={result.party.id}
									className="flex items-center p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
								>
									<div className="flex items-center flex-1 min-w-0">
										<div className="flex-shrink-0 mr-4">
											<div className="w-3 h-3 rounded-full flex items-center justify-center text-white text-xs font-bold">
												<span
													className={`w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-bold ${getPartyColorClass(result.party.shortName)}`}
												>
													{result.party.shortName}
												</span>
											</div>
										</div>
										<div className="flex-1 min-w-0">
											<h3 className="font-semibold text-gray-900 truncate">
												{result.party.name}
											</h3>
											<p className="text-sm text-gray-600">
												{result.matchingVotes} van {result.totalVotes}{" "}
												stellingen overeenkomen
											</p>
										</div>
									</div>
									<div className="flex items-center ml-4">
										<div className="text-right mr-4">
											<div className="font-bold text-lg">
												{formatPercentage(result.agreement)}
											</div>
											<div className="text-xs text-gray-500">
												Score: {result.score.toFixed(1)}
											</div>
										</div>
										<div className="w-24">
											<Progress
												value={result.agreement}
												size="sm"
												variant={index === 0 ? "success" : "default"}
											/>
										</div>
									</div>
								</div>
							))}
						</div>

						{!showAllParties && otherResults.length > 0 && (
							<div className="text-center mt-6">
								<Button
									variant="secondary"
									onClick={() => setShowAllParties(true)}
								>
									Toon alle {results.partyResults.length} partijen
								</Button>
							</div>
						)}

						{showAllParties && otherResults.length > 0 && (
							<div className="text-center mt-6">
								<Button
									variant="secondary"
									onClick={() => setShowAllParties(false)}
								>
									Toon alleen top 3
								</Button>
							</div>
						)}
					</CardContent>
				</Card>

				{/* Motion Details */}
				{results.motionDetails && results.motionDetails.length > 0 && (
					<Card className="mb-8">
						<CardHeader>
							<CardTitle>Gedetailleerde uitslagen</CardTitle>
							<CardDescription>
								Bekijk hoe de partijen hebben gestemd op elke stelling en
								vergelijk met jouw antwoord.
							</CardDescription>
							<div className="mt-3 p-3 bg-blue-50 border border-blue-200 rounded-lg">
								<p className="text-sm text-blue-800">
									💡 <strong>Let op:</strong> Neutrale antwoorden worden niet
									meegeteld in de berekening en zijn grijs weergegeven.
								</p>
							</div>
						</CardHeader>
						<CardContent>
							<div className="space-y-4">
								{results.motionDetails.map(
									(detail: MotionDetail, index: number) => (
										<div
											key={detail.motionId}
											className={`border rounded-lg ${
												detail.userAnswer === "neutral"
													? "border-gray-300 bg-gray-50 opacity-75"
													: "border-gray-200"
											}`}
										>
											<button
												type="button"
												onClick={() => toggleMotionDetails(detail.motionId)}
												className="w-full p-4 text-left hover:bg-gray-50 transition-colors flex items-center justify-between"
											>
												<div className="flex items-center space-x-3">
													<span className="text-2xl">
														{getUserAnswerEmoji(detail.userAnswer)}
													</span>
													<div>
														<h3 className="font-semibold text-gray-900">
															Stelling {index + 1}:{" "}
															{detail.motion?.title || "Untitled"}
														</h3>
														<p className="text-sm text-gray-600 mt-1">
															Jij antwoordde:{" "}
															<span
																className={`font-medium ${
																	detail.userAnswer === "neutral"
																		? "text-gray-500"
																		: ""
																}`}
															>
																{detail.userAnswer === "agree"
																	? "Eens"
																	: detail.userAnswer === "disagree"
																		? "Oneens"
																		: "Neutraal (niet meegeteld)"}
															</span>
														</p>
													</div>
												</div>
												{expandedMotions.has(detail.motionId) ? (
													<ChevronUp className="h-5 w-5 text-gray-400" />
												) : (
													<ChevronDown className="h-5 w-5 text-gray-400" />
												)}
											</button>

											{expandedMotions.has(detail.motionId) && (
												<div className="border-t border-gray-200 p-4 bg-gray-50">
													{detail.motion?.description && (
														<p className="text-gray-700 mb-4">
															{detail.motion.description}
														</p>
													)}

													{detail.motion?.bulletPoints &&
														detail.motion.bulletPoints.length > 0 && (
															<div className="mb-4">
																<h4 className="font-medium text-gray-900 mb-2">
																	Kernpunten:
																</h4>
																<ul className="space-y-1">
																	{detail.motion.bulletPoints.map(
																		(point, pointIndex) => (
																			<li
																				key={`motion-${detail.motionId}-bullet-${pointIndex}`}
																				className="text-sm text-gray-700 flex items-start"
																			>
																				<span className="text-primary-500 mr-2 mt-0.5">
																					•
																				</span>
																				{point}
																			</li>
																		),
																	)}
																</ul>
															</div>
														)}

													<h4 className="font-medium text-gray-900 mb-3">
														Hoe stemden de partijen?
													</h4>
													<div className="grid gap-2">
														{detail.partyPositions
															.sort((a, b) => {
																// Sort by agreement with user first, then by party name
																if (a.agreesWithUser !== b.agreesWithUser) {
																	return a.agreesWithUser ? -1 : 1;
																}
																return a.party.shortName.localeCompare(
																	b.party.shortName,
																);
															})
															.map((partyPos) => (
																<div
																	key={partyPos.party.id}
																	className={`flex items-center justify-between p-3 rounded-lg border ${
																		detail.userAnswer === "neutral"
																			? "bg-gray-50 border-gray-200 opacity-60"
																			: partyPos.agreesWithUser
																				? "bg-green-50 border-green-200"
																				: "bg-red-50 border-red-200"
																	}`}
																>
																	<div className="flex items-center space-x-3">
																		<span className="text-lg">
																			{getVoteEmoji(partyPos.position)}
																		</span>
																		<div>
																			<span className="font-medium text-gray-900">
																				{partyPos.party.shortName}
																			</span>
																			<span className="text-sm text-gray-600 ml-2">
																				({partyPos.party.name})
																			</span>
																		</div>
																	</div>
																	<div className="flex items-center space-x-2">
																		<span className="text-sm text-gray-600">
																			{partyPos.position === "FOR"
																				? "Voor"
																				: partyPos.position === "AGAINST"
																					? "Tegen"
																					: "Onthouding"}
																		</span>
																		{detail.userAnswer === "neutral" ? (
																			<span className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded-full">
																				Niet meegeteld
																			</span>
																		) : partyPos.agreesWithUser ? (
																			<span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded-full">
																				✓ Eens met jou
																			</span>
																		) : null}
																	</div>
																</div>
															))}
													</div>
												</div>
											)}
										</div>
									),
								)}
							</div>
						</CardContent>
					</Card>
				)}

				{/* What's Next */}
				<div className="grid md:grid-cols-2 gap-6">
					<Card>
						<CardHeader>
							<CardTitle>Meer stellingen beantwoorden</CardTitle>
							<CardDescription>
								Wil je een nauwkeuriger resultaat? Beantwoord meer stellingen om
								je politieke voorkeur nog beter in kaart te brengen.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<Link to={`/compass?session=${sessionId}`} className="block">
								<Button className="w-full">
									<ArrowRight className="h-4 w-4 mr-2" />
									Meer stellingen
								</Button>
							</Link>
						</CardContent>
					</Card>

					<Card>
						<CardHeader>
							<CardTitle>Begin opnieuw</CardTitle>
							<CardDescription>
								Nieuwsgierig naar andere onderwerpen? Start een nieuwe
								stemwijzer met verschillende stellingen.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<Link to="/compass" className="block">
								<Button variant="secondary" className="w-full">
									<RotateCcw className="h-4 w-4 mr-2" />
									Nieuwe stemwijzer
								</Button>
							</Link>
						</CardContent>
					</Card>
				</div>

				{/* Disclaimer */}
				<div className="mt-8 p-6 bg-yellow-50 border border-yellow-200 rounded-lg">
					<h3 className="font-semibold text-yellow-800 mb-2">
						💡 Belangrijk om te weten
					</h3>
					<p className="text-sm text-yellow-700 leading-relaxed">
						Deze resultaten zijn gebaseerd op daadwerkelijke stemmingen in de
						Tweede Kamer, maar geven slechts een indicatie. Partijstandpunten
						kunnen evolueren en individuele politici binnen een partij kunnen
						soms anders stemmen. Gebruik deze tool als startpunt voor verder
						onderzoek naar partijprogramma's en standpunten.
					</p>
				</div>
			</div>
		</div>
	);
}
