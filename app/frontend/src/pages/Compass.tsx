import {
	ArrowLeft,
	ArrowRight,
	CheckCircle,
	ChevronDown,
	ExternalLink,
	Meh,
	RotateCcw,
	ThumbsDown,
	ThumbsUp,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { Button } from "../components/ui/Button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../components/ui/Card";
import { Progress } from "../components/ui/Progress";
import {
	useCompassMotions,
	useCompassMotionsCount,
	useCompassResults,
	useSubmitAnswers,
} from "../hooks/api";
import type { Motion, UserAnswer } from "../lib/api";
import { saveSessionId } from "../lib/sessionStorage";
import { calculateProgress } from "../lib/utils";

type Answer = "agree" | "disagree" | "neutral";
const COMPASS_QUESTION_COUNT = 20;

interface CompassState {
	currentIndex: number;
	answers: UserAnswer[];
	showExplanation: boolean;
}

const formatTitle = (title: string | undefined): string => {
	if (!title) {
		return "";
	}
	const cleanedTitle = title
		.replace(/^motie van (het lid|de leden) .*? over /i, "")
		.trim();
	return cleanedTitle.charAt(0).toUpperCase() + cleanedTitle.slice(1);
};

export function CompassPage() {
	const navigate = useNavigate();
	const [searchParams] = useSearchParams();
	const sessionId = searchParams.get("session");
	const [state, setState] = useState<CompassState>({
		currentIndex: 0,
		answers: [],
		showExplanation: false,
	});

	const [motions, setMotions] = useState<Motion[]>([]);
	const [showAllBulletPoints, setShowAllBulletPoints] = useState(false);

	// Get existing results if continuing a session
	const { data: existingResults } = useCompassResults(sessionId || "");

	// Parse query parameters for filtering
	const filterParams = useMemo(() => {
		const categoryIds = searchParams
			.get("categoryIds")
			?.split(",")
			.filter(Boolean);
		const afterParam = searchParams.get("after");
		const after = afterParam ? new Date(afterParam) : undefined;
		const search = searchParams.get("search") || undefined;
		const partyIds = searchParams
			.get("partyIds")
			?.split(",")
			.filter(Boolean);

		return {
			categoryIds,
			after,
			search,
			partyIds,
		};
	}, [searchParams]);

	const submitAnswers = useSubmitAnswers();

	const { data: motionCount, isLoading: isCountLoading } =
		useCompassMotionsCount(filterParams);

	const {
		data: newMotions,
		isLoading,
		isFetching,
		isError,
	} = useCompassMotions(
		COMPASS_QUESTION_COUNT,
		motions.map((m) => m.id),
		filterParams.categoryIds,
		filterParams.after,
		filterParams.search,
		filterParams.partyIds,
	);

	useEffect(() => {
		if (newMotions) {
			setMotions((prev) => {
				const existingIds = new Set(prev.map((m) => m.id));
				const uniqueNew = newMotions.filter((m) => !existingIds.has(m.id));
				return [...prev, ...uniqueNew];
			});
		}
	}, [newMotions]);

	// Load existing answers when continuing a session
	useEffect(() => {
		if (existingResults?.motionDetails) {
			const existingAnswers: UserAnswer[] = existingResults.motionDetails.map(
				(detail) => ({
					motionId: detail.motionId,
					answer: detail.userAnswer,
				}),
			);

			setState((prev) => ({
				...prev,
				answers: existingAnswers,
			}));
		}
	}, [existingResults]);

	const getUnansweredMotions = () => {
		if (!sessionId) return motions;
		const answeredMotionIds = new Set(state.answers.map((a) => a.motionId));
		return motions.filter((motion) => !answeredMotionIds.has(motion.id));
	};

	const unansweredMotions = getUnansweredMotions();
	const currentMotion = sessionId
		? unansweredMotions[state.currentIndex]
		: motions[state.currentIndex];

	const totalAvailableMotions = motionCount?.count ?? 0;

	const questionsToAnswer = Math.min(
		COMPASS_QUESTION_COUNT,
		totalAvailableMotions,
	);

	const questionNumber = Math.min(
		(existingResults?.motionDetails?.length ?? 0) + state.currentIndex + 1,
		questionsToAnswer,
	);

	// Progress through ALL loaded motions
	const progress = calculateProgress(state.answers.length, questionsToAnswer);

	// Minimum answers needed is the smaller of 20 or total available motions
	const minAnswersRequired = Math.min(
		COMPASS_QUESTION_COUNT,
		totalAvailableMotions,
	);

	const displayedBulletPoints = useMemo(() => {
		const allPoints = currentMotion?.bulletPoints || [];
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
	}, [currentMotion, showAllBulletPoints]);

	const hasAdvisoryPoints = useMemo(() => {
		const allPoints = currentMotion?.bulletPoints || [];
		return allPoints.some((p) =>
			p.toLowerCase().trimStart().startsWith("verzoekt"),
		);
	}, [currentMotion]);

	const handleAnswer = (answer: Answer) => {
		if (!currentMotion) return;

		const newAnswer: UserAnswer = {
			motionId: currentMotion.id,
			answer,
		};

		const updatedAnswers = [...state.answers];
		const existingIndex = updatedAnswers.findIndex(
			(a) => a.motionId === currentMotion.id,
		);

		if (existingIndex >= 0) {
			updatedAnswers[existingIndex] = newAnswer;
		} else {
			updatedAnswers.push(newAnswer);
		}

		setState((prev) => ({ ...prev, answers: updatedAnswers }));

		// Auto-advance to next question after a short delay
		setTimeout(() => {
			setShowAllBulletPoints(false); // Reset bullet points view for next question
			if (state.currentIndex < motions.length - 1) {
				setState((prev) => ({
					...prev,
					currentIndex: prev.currentIndex + 1,
				}));
			}
		}, 300);
	};

	const handleSubmit = async (answers: UserAnswer[]) => {
		try {
			const result = await submitAnswers.mutateAsync({ answers });
			// Save session ID to localStorage for later retrieval
			saveSessionId(result.id);
			
			// Build URL with query parameters
			const params = new URLSearchParams();
			if (filterParams.categoryIds?.length) {
				params.set("categoryIds", filterParams.categoryIds.join(","));
			}
			if (filterParams.after) {
				params.set("after", filterParams.after.toISOString().split("T")[0]);
			}
			if (filterParams.search) {
				params.set("search", filterParams.search);
			}
			if (filterParams.partyIds?.length) {
				params.set("partyIds", filterParams.partyIds.join(","));
			}
			
			const queryString = params.toString();
			navigate(`/results/${result.id}${queryString ? `?${queryString}` : ""}`);
		} catch (error) {
			console.error("Failed to submit answers:", error);
		}
	};

	const handleReset = () => {
		setState({
			currentIndex: 0,
			answers: sessionId ? [] : [],
			showExplanation: false,
		});
		setMotions([]); // Clear motions for a fresh start
		if (sessionId) {
			navigate("/compass");
		}
	};

	const getCurrentAnswer = (): Answer | undefined => {
		return state.answers.find((a) => a.motionId === currentMotion?.id)?.answer;
	};

	if (isError) {
		return <div>Error loading questions...</div>;
	}

	if (
		(isLoading && motions.length === 0) ||
		(isFetching && !currentMotion) ||
		isCountLoading
	) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<div className="text-center">
					<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
					<p className="text-gray-600">Stellingen laden...</p>
				</div>
			</div>
		);
	}

	if (totalAvailableMotions === 0 && !isLoading && !isFetching) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<Card className="max-w-md mx-auto">
					<CardHeader>
						<CardTitle>Geen moties gevonden</CardTitle>
						<CardDescription>
							Er zijn geen moties gevonden die voldoen aan je filters. Pas je
							instellingen aan om meer moties te zien.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<Link to="/compass/settings">
							<Button variant="primary" className="w-full">
								Terug naar instellingen
							</Button>
						</Link>
					</CardContent>
				</Card>
			</div>
		);
	}

	return (
		<div className="bg-gray-50 pb-48">
			{/* Warning banner for limited motions */}
			{totalAvailableMotions < COMPASS_QUESTION_COUNT && (
				<div className="bg-yellow-50 border-b border-yellow-200">
					<div className="container mx-auto px-4 py-3">
						<p className="text-sm text-yellow-800 text-center">
							⚠️ Let op: Door je filters zijn er maar{" "}
							<strong>{totalAvailableMotions} moties</strong> beschikbaar. Overweeg
							je filters aan te passen voor meer stellingen.{" "}
							<Link
								to="/compass/settings"
								className="underline font-medium hover:text-yellow-900"
							>
								Wijzig filters
							</Link>
						</p>
					</div>
				</div>
			)}

			{/* --- Sticky Header --- */}
			<div className="sticky top-0 bg-white shadow-sm z-20">
				<div className="container mx-auto px-4 py-3">
					<div className="flex items-center justify-between gap-2 md:gap-4">
						<Link
							to="/compass/settings"
							className="flex items-center text-sm text-gray-600 hover:text-gray-900 whitespace-nowrap"
						>
							<ArrowLeft className="h-4 w-4 mr-1 md:mr-2" />
							<span className="hidden sm:inline">Instellingen</span>
						</Link>
						<div className="flex-grow max-w-lg">
							<Progress value={progress} />
							<div className="text-xs text-gray-500 mt-1 text-center">
								Vraag {questionNumber} van {questionsToAnswer}
							</div>
						</div>
						<Button
							variant="ghost"
							onClick={handleReset}
							className="text-gray-500 hover:text-gray-700 flex-shrink-0"
							aria-label="Reset compass"
							size="sm"
						>
							<RotateCcw className="h-4 w-4" />
						</Button>
					</div>
				</div>
			</div>

			{/* --- Main Content --- */}
			<div className="container mx-auto px-4 py-6">
				<div className="max-w-4xl mx-auto">
					<Card
						className={`mb-6 transition-all ${getCurrentAnswer() ? "ring-2 ring-green-200 bg-green-50/50" : ""}`}
					>
						<CardHeader>
							<div className="flex items-start justify-between">
								<CardTitle className="text-2xl leading-tight">
									{formatTitle(currentMotion?.title)}
								</CardTitle>
								{currentMotion?.motionNumber && currentMotion.did && (
									<a
										href={`https://www.tweedekamer.nl/kamerstukken/moties/detail?id=${currentMotion.motionNumber}&did=${currentMotion.did}`}
										target="_blank"
										rel="noopener noreferrer"
										className="btn focus-ring btn-secondary btn-sm ml-4 flex-shrink-0 inline-flex items-center"
									>
										<ExternalLink className="h-4 w-4 mr-2" />
										Bron
									</a>
								)}
							</div>
							{currentMotion?.description && (
								<CardDescription className="text-sm text-gray-600 leading-relaxed pt-2">
									{currentMotion.description}
								</CardDescription>
							)}
						</CardHeader>

						{currentMotion?.bulletPoints &&
							currentMotion.bulletPoints.length > 0 && (
								<CardContent className="pt-0">
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
														key={`bullet-${currentMotion.id}-${index}`}
														className={`text-sm flex items-start ${
															isAdvisoryPoint
																? "text-primary-900 font-semibold bg-primary-50/70 rounded-md p-2 border-l-3 border-primary-400"
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
										{currentMotion.bulletPoints.length >
											displayedBulletPoints.length && (
											<Button
												onClick={() =>
													setShowAllBulletPoints(!showAllBulletPoints)
												}
												className="p-0 h-auto mt-3 text-sm bg-transparent text-gray-800 hover:bg-transparent"
											>
												{!showAllBulletPoints && (
													<>
														<ChevronDown className="h-4 w-4 mr-1" />
														Toon motivatie (
														{currentMotion.bulletPoints.length -
															displayedBulletPoints.length}
														)
													</>
												)}
											</Button>
										)}
									</div>
								</CardContent>
							)}
					</Card>
				</div>
			</div>

			{/* --- Sticky Footer --- */}
			<div className="fixed bottom-0 left-0 right-0 bg-white/80 backdrop-blur-sm border-t border-gray-200 z-10">
				<div className="container mx-auto px-4 py-4">
					<div className="max-w-4xl mx-auto">
						{/* Results CTA - always visible, prominent when ready */}
						{state.answers.length > 0 && (
							<div className="mb-4 p-3 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-200">
								<div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
									<div className="text-sm font-medium text-gray-700">
										<div className="flex items-center gap-2">
											<span className="font-semibold text-blue-600 text-base">
												{state.answers.length}/{minAnswersRequired}
											</span>
											<span className="hidden sm:inline">beantwoord</span>
										</div>
										<div className="text-xs text-gray-500 mt-0.5">
											{minAnswersRequired < COMPASS_QUESTION_COUNT ? (
												<>
													Minimaal {minAnswersRequired} voor resultaten (beperkte
													moties beschikbaar)
												</>
											) : (
												<>Minimaal {minAnswersRequired} voor resultaten</>
											)}
										</div>
									</div>
									<Button
										onClick={() => handleSubmit(state.answers)}
										disabled={
											state.answers.length < minAnswersRequired ||
											submitAnswers.isPending
										}
										loading={submitAnswers.isPending}
										className={`flex-shrink-0 transition-all w-full sm:w-auto ${
											state.answers.length >= minAnswersRequired
												? "bg-blue-600 text-white hover:bg-blue-700 shadow-md hover:shadow-lg"
												: "bg-gray-200 text-gray-400 cursor-not-allowed"
										}`}
										size="sm"
									>
										Bekijk resultaten
										<ArrowRight className="h-4 w-4 ml-2" />
									</Button>
								</div>
							</div>
						)}

						<div className="grid grid-cols-3 gap-2 sm:gap-3 mb-4">
							<Button
								variant={
									getCurrentAnswer() === "agree" ? "primary" : "secondary"
								}
								size="lg"
								onClick={() => handleAnswer("agree")}
								className="h-16 sm:h-20 text-base sm:text-lg font-semibold relative"
								disabled={submitAnswers.isPending}
							>
								{getCurrentAnswer() === "agree" && (
									<CheckCircle className="absolute top-1 right-1 sm:top-2 sm:right-2 h-4 w-4 sm:h-5 sm:w-5" />
								)}
								<div className="flex flex-col items-center">
									<ThumbsUp className="h-6 w-6 sm:h-8 sm:w-8 mb-0.5 sm:mb-1" />
									<span className="text-xs sm:text-base">Eens</span>
								</div>
							</Button>

							<Button
								variant={
									getCurrentAnswer() === "neutral" ? "primary" : "secondary"
								}
								size="lg"
								onClick={() => handleAnswer("neutral")}
								className="h-16 sm:h-20 text-base sm:text-lg font-semibold relative"
								disabled={submitAnswers.isPending}
							>
								{getCurrentAnswer() === "neutral" && (
									<CheckCircle className="absolute top-1 right-1 sm:top-2 sm:right-2 h-4 w-4 sm:h-5 sm:w-5" />
								)}
								<div className="flex flex-col items-center">
									<Meh className="h-6 w-6 sm:h-8 sm:w-8 mb-0.5 sm:mb-1" />
									<span className="text-xs sm:text-base">Neutraal</span>
								</div>
							</Button>

							<Button
								variant={
									getCurrentAnswer() === "disagree" ? "primary" : "secondary"
								}
								size="lg"
								onClick={() => handleAnswer("disagree")}
								className="h-16 sm:h-20 text-base sm:text-lg font-semibold relative"
								disabled={submitAnswers.isPending}
							>
								{getCurrentAnswer() === "disagree" && (
									<CheckCircle className="absolute top-1 right-1 sm:top-2 sm:right-2 h-4 w-4 sm:h-5 sm:w-5" />
								)}
								<div className="flex flex-col items-center">
									<ThumbsDown className="h-6 w-6 sm:h-8 sm:w-8 mb-0.5 sm:mb-1" />
									<span className="text-xs sm:text-base">Oneens</span>
								</div>
							</Button>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}
