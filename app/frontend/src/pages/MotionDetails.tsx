import {
	ArrowLeft,
	Calendar,
	ChevronDown,
	ExternalLink,
	FileText,
	Users,
	Vote,
} from "lucide-react";
import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Button } from "../components/ui/Button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../components/ui/Card";
import { useMotion, useMotionDetails } from "../hooks/api";
import type { VoteType } from "../lib/api";
import { formatDate, getPartyColorClass } from "../lib/utils";

export const MotionDetailsPage = () => {
	const { motionId } = useParams<{ motionId: string }>();
	const [showAllBulletPoints, setShowAllBulletPoints] = useState(false);

	const {
		data: motion,
		isLoading: isMotionLoading,
		error: motionError,
	} = useMotion(motionId || "");

	const {
		data: motionDetails,
		isLoading: isDetailsLoading,
		error: detailsError,
	} = useMotionDetails(motionId || "", true);

	const displayedBulletPoints = useMemo(() => {
		const allPoints = motion?.bulletPoints || [];
		const advisoryPoints = allPoints.filter((p) =>
			p.toLowerCase().trimStart().startsWith("verzoekt"),
		);

		if (showAllBulletPoints) {
			if (advisoryPoints.length > 0) {
				const otherPoints = allPoints.filter(
					(p) => !p.toLowerCase().trimStart().startsWith("verzoekt"),
				);
				return [...advisoryPoints, ...otherPoints];
			}
			return allPoints;
		}

		return advisoryPoints.length > 0 ? advisoryPoints : allPoints;
	}, [motion, showAllBulletPoints]);

	const hasAdvisoryPoints = useMemo(() => {
		const allPoints = motion?.bulletPoints || [];
		return allPoints.some((p) =>
			p.toLowerCase().trimStart().startsWith("verzoekt"),
		);
	}, [motion]);

	if (isMotionLoading || isDetailsLoading) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<div className="text-center">
					<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
					<p className="text-gray-600">Motie laden...</p>
				</div>
			</div>
		);
	}

	if (motionError || detailsError || !motion) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<Card className="max-w-md mx-auto">
					<CardHeader>
						<CardTitle>Motie niet gevonden</CardTitle>
						<CardDescription>
							De motie kon niet worden geladen. Mogelijk bestaat deze niet meer.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<Link to="/">
							<button
								type="button"
								className="w-full bg-primary-600 text-white px-4 py-2 rounded-md hover:bg-primary-700 transition-colors"
							>
								Terug naar home
							</button>
						</Link>
					</CardContent>
				</Card>
			</div>
		);
	}

	const getVoteLabel = (voteType: VoteType) => {
		switch (voteType) {
			case "FOR":
				return "Voor";
			case "AGAINST":
				return "Tegen";
			case "NEUTRAL":
				return "Onthouding";
			default:
				return "Onbekend";
		}
	};

	const partyPositions = motionDetails?.partyPositions || [];
	const positionsByType = {
		FOR: partyPositions.filter((p) => p.position === "FOR"),
		AGAINST: partyPositions.filter((p) => p.position === "AGAINST"),
		NEUTRAL: partyPositions.filter((p) => p.position === "NEUTRAL"),
	};

	const totalVotes = partyPositions.reduce((sum, p) => sum + p.count, 0);
	const forVotes = positionsByType.FOR.reduce((sum, p) => sum + p.count, 0);
	const againstVotes = positionsByType.AGAINST.reduce(
		(sum, p) => sum + p.count,
		0,
	);
	const neutralVotes = positionsByType.NEUTRAL.reduce(
		(sum, p) => sum + p.count,
		0,
	);

	const forPercentage = totalVotes > 0 ? (forVotes / totalVotes) * 100 : 0;
	const againstPercentage =
		totalVotes > 0 ? (againstVotes / totalVotes) * 100 : 0;
	const neutralPercentage =
		totalVotes > 0 ? (neutralVotes / totalVotes) * 100 : 0;

	return (
		<div className="min-h-screen bg-gray-50">
			<div className="container mx-auto px-4 py-8 max-w-4xl">
				{/* Navigation */}
				<div className="mb-6">
					<Link
						to="/motions"
						className="inline-flex items-center text-gray-600 hover:text-gray-800 transition-colors"
					>
						<ArrowLeft className="h-4 w-4 mr-2" />
						Terug naar moties
					</Link>
				</div>

				{/* Motion Header */}
				<Card className="mb-6">
					<CardHeader>
						<div className="flex items-start justify-between">
							<div className="flex-1">
								<CardTitle className="text-2xl leading-tight">
									{motion.title}
								</CardTitle>
								{motion.description && (
									<CardDescription className="text-sm text-gray-600 leading-relaxed pt-2">
										{motion.description}
									</CardDescription>
								)}
								{motion.shortTitle && motion.shortTitle !== motion.title && (
									<CardDescription className="text-sm text-gray-600 pt-2">
										{motion.shortTitle}
									</CardDescription>
								)}
							</div>
							{motion.motionNumber && motion.did && (
								<a
									href={`https://www.tweedekamer.nl/kamerstukken/moties/detail?id=${motion.motionNumber}&did=${motion.did}`}
									target="_blank"
									rel="noopener noreferrer"
									className="btn focus-ring btn-secondary btn-sm ml-4 flex-shrink-0 inline-flex items-center"
								>
									<ExternalLink className="h-4 w-4 mr-2" />
									Bron
								</a>
							)}
						</div>
					</CardHeader>

					<CardContent className="pt-0">
						{motion.bulletPoints && motion.bulletPoints.length > 0 ? (
							<div className="bg-gray-50/80 rounded-lg p-4">
								{hasAdvisoryPoints && (
									<div className="mb-3">
										<h4 className="text-sm font-semibold text-primary-700 mb-2 flex items-center">
											<span className="inline-block w-2 h-2 bg-primary-500 rounded-full mr-2"></span>
											Kernverzoek
										</h4>
									</div>
								)}
								<ul className="space-y-1.5">
									{displayedBulletPoints.map((point, index) => {
										const isAdvisoryPoint = point
											.toLowerCase()
											.trimStart()
											.startsWith("verzoekt");
										return (
											<li
												key={`bullet-${motion.id}-${index}`}
												className={`text-sm flex items-start ${
													isAdvisoryPoint
														? "text-primary-900 font-semibold bg-primary-50/70 rounded-md p-2 border-l-4 border-primary-400"
														: "text-gray-800"
												}`}
											>
												<span
													className={`mr-2.5 mt-0.5 ${
														isAdvisoryPoint
															? "text-primary-600"
															: "text-primary-500"
													}`}
												>
													{isAdvisoryPoint ? "→" : "•"}
												</span>
												{point}
											</li>
										);
									})}
								</ul>
								{motion.bulletPoints.length > displayedBulletPoints.length && (
									<Button
										onClick={() => setShowAllBulletPoints(!showAllBulletPoints)}
										className="p-0 h-auto mt-3 text-sm bg-transparent text-gray-800 hover:bg-transparent"
									>
										{!showAllBulletPoints && (
											<>
												<ChevronDown className="h-4 w-4 mr-1" />
												Toon motivatie (
												{motion.bulletPoints.length -
													displayedBulletPoints.length}
												)
											</>
										)}
									</Button>
								)}
							</div>
						) : (
							<div className="bg-yellow-50/80 rounded-lg p-4 border-l-4 border-yellow-400">
								<div className="flex items-start">
									<span className="text-yellow-600 mr-2 mt-0.5">⚠️</span>
									<div>
										<h4 className="text-sm font-semibold text-yellow-800 mb-1">
											Inhoud nog niet beschikbaar
										</h4>
										<p className="text-sm text-yellow-700">
											De gedetailleerde inhoud van deze motie is nog niet
											verwerkt. Dit betreft waarschijnlijk een recent ingediende
											motie.
										</p>
									</div>
								</div>
							</div>
						)}

						{/* Motion metadata below content */}
						<div className="flex flex-wrap gap-4 text-sm text-gray-500 mt-4 pt-4 border-t border-gray-200">
							{motion.motionNumber && (
								<div className="flex items-center">
									<FileText className="h-4 w-4 mr-1" />
									{motion.motionNumber}
								</div>
							)}
							{motion.createdAt && (
								<div className="flex items-center">
									<Calendar className="h-4 w-4 mr-1" />
									{formatDate(motion.createdAt)}
								</div>
							)}
							{motion.status && (
								<div className="flex items-center">
									<Vote className="h-4 w-4 mr-1" />
									{motion.status}
								</div>
							)}
						</div>
					</CardContent>
				</Card>

				{/* Voting Results */}
				<Card className="mb-6">
					<CardHeader>
						<CardTitle className="flex items-center">
							<Users className="h-6 w-6 mr-3 text-primary-600" />
							Stemresultaten
						</CardTitle>
						<CardDescription>
							Hoe hebben de verschillende partijen gestemd op deze motie?
						</CardDescription>
					</CardHeader>
					<CardContent>
						{partyPositions.length > 0 ? (
							<>
								{/* Vote Summary */}
								<div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
									<div className="bg-green-50 p-4 rounded-lg">
										<div className="flex items-center justify-between mb-2">
											<span className="text-green-800 font-medium">Voor</span>
											<div className="w-3 h-3 bg-green-500 rounded-full"></div>
										</div>
										<div className="text-2xl font-bold text-green-900">
											{forVotes}
										</div>
										<div className="text-sm text-green-700">
											{Math.round(forPercentage)}% van de stemmen
										</div>
									</div>

									<div className="bg-red-50 p-4 rounded-lg">
										<div className="flex items-center justify-between mb-2">
											<span className="text-red-800 font-medium">Tegen</span>
											<div className="w-3 h-3 bg-red-500 rounded-full"></div>
										</div>
										<div className="text-2xl font-bold text-red-900">
											{againstVotes}
										</div>
										<div className="text-sm text-red-700">
											{Math.round(againstPercentage)}% van de stemmen
										</div>
									</div>

									<div className="bg-yellow-50 p-4 rounded-lg">
										<div className="flex items-center justify-between mb-2">
											<span className="text-yellow-800 font-medium">
												Onthouding
											</span>
											<div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
										</div>
										<div className="text-2xl font-bold text-yellow-900">
											{neutralVotes}
										</div>
										<div className="text-sm text-yellow-700">
											{Math.round(neutralPercentage)}% van de stemmen
										</div>
									</div>
								</div>

								{/* Detailed Votes by Party */}
								<div className="space-y-4">
									<h3 className="font-semibold text-lg">
										Stemming per partij:
									</h3>

									{Object.entries(positionsByType).map(
										([voteType, partiesWithPosition]) =>
											partiesWithPosition.length > 0 && (
												<div key={voteType} className="mb-6">
													<h4 className="font-medium mb-3 flex items-center">
														<div
															className={`w-3 h-3 mr-2 rounded-full ${
																voteType === "FOR"
																	? "bg-green-500"
																	: voteType === "AGAINST"
																		? "bg-red-500"
																		: "bg-yellow-500"
															}`}
														></div>
														{getVoteLabel(voteType as VoteType)} (
														{partiesWithPosition.reduce(
															(sum, p) => sum + p.count,
															0,
														)}{" "}
														stemmen)
													</h4>
													<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
														{partiesWithPosition.map((partyPosition) => (
															<div
																key={partyPosition.party.id}
																className="flex items-center justify-between p-3 bg-white rounded-md border border-gray-200"
															>
																<div className="flex items-center">
																	<div
																		className={`w-4 h-4 rounded-full mr-3 ${getPartyColorClass(partyPosition.party.shortName)}`}
																	></div>
																	<div>
																		<div className="font-medium text-sm">
																			{partyPosition.party.shortName}
																		</div>
																	</div>
																</div>
																<div className="text-sm font-semibold text-gray-700">
																	{partyPosition.count}
																</div>
															</div>
														))}
													</div>
												</div>
											),
									)}
								</div>
							</>
						) : (
							<div className="bg-blue-50/80 rounded-lg p-4 border-l-4 border-blue-400">
								<div className="flex items-start">
									<span className="text-blue-600 mr-2 mt-0.5">ℹ️</span>
									<div>
										<h4 className="text-sm font-semibold text-blue-800 mb-1">
											Stemresultaten nog niet beschikbaar
										</h4>
										<p className="text-sm text-blue-700">
											Er zijn nog geen stemresultaten bekend voor deze motie.
											Dit kan betekenen dat de motie nog niet in behandeling is
											genomen of dat de stemmingsgegevens nog niet zijn
											verwerkt.
										</p>
									</div>
								</div>
							</div>
						)}
					</CardContent>
				</Card>

				{/* Categories */}
				{motion.categories && motion.categories.length > 0 && (
					<Card className="mb-6">
						<CardHeader>
							<CardTitle>Categorieën</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="flex flex-wrap gap-2">
								{motion.categories.map((category) => (
									<span
										key={category.id}
										className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-800"
									>
										{category.name}
									</span>
								))}
							</div>
						</CardContent>
					</Card>
				)}

				{/* Additional Information */}
				{(motion.createdAt || motion.updatedAt || totalVotes > 0) && (
					<Card>
						<CardHeader>
							<CardTitle>Aanvullende informatie</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
								{motion.createdAt && (
									<div>
										<span className="font-medium text-gray-600">
											Ingediend:
										</span>
										<span className="ml-2">{formatDate(motion.createdAt)}</span>
									</div>
								)}
								{motion.updatedAt && motion.updatedAt !== motion.createdAt && (
									<div>
										<span className="font-medium text-gray-600">
											Laatst gewijzigd:
										</span>
										<span className="ml-2">{formatDate(motion.updatedAt)}</span>
									</div>
								)}
								{totalVotes > 0 && (
									<div>
										<span className="font-medium text-gray-600">
											Totaal stemmen:
										</span>
										<span className="ml-2">{totalVotes}</span>
									</div>
								)}
							</div>
						</CardContent>
					</Card>
				)}
			</div>
		</div>
	);
};
