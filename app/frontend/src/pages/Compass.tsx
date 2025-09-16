import {
	ArrowLeft,
	ArrowRight,
	CheckCircle,
	ChevronDown,
	ChevronUp,
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
	useCompassResults,
	useSubmitAnswers,
} from "../hooks/api";
import type { Motion, UserAnswer } from "../lib/api";
import { calculateProgress } from "../lib/utils";

type Answer = "agree" | "disagree" | "neutral";

interface CompassState {
	currentIndex: number;
	answers: UserAnswer[];
	showExplanation: boolean;
}

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
	const [hasMore, setHasMore] = useState(true);
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

		return {
			categoryIds,
			after,
		};
	}, [searchParams]);

	const submitAnswers = useSubmitAnswers();

	const {
		data: newMotions,
		isLoading,
		isFetching,
		isError,
	} = useCompassMotions(
		20,
		motions.map((m) => m.id),
		filterParams.categoryIds,
		filterParams.after,
	);

	useEffect(() => {
		if (newMotions) {
			if (newMotions.length < 20) {
				setHasMore(false);
			}
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
	const totalMotions = sessionId ? unansweredMotions.length : motions.length;
	const progress = calculateProgress(state.currentIndex + 1, totalMotions);
	const isLastQuestion = !hasMore && state.currentIndex === totalMotions - 1;

	const displayedBulletPoints = useMemo(() => {
		const allPoints = currentMotion?.bulletPoints || [];
		if (showAllBulletPoints) {
			return allPoints;
		}
		const verzoektPoints = allPoints.filter((p) =>
			p.toLowerCase().trimStart().startsWith("verzoekt"),
		);
		return verzoektPoints.length > 0 ? verzoektPoints : allPoints;
	}, [currentMotion, showAllBulletPoints]);

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
			navigate(`/results/${result.id}`);
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
		setHasMore(true);
		if (sessionId) {
			navigate("/compass");
		}
	};

	const getCurrentAnswer = (): Answer | undefined => {
		return state.answers.find((a) => a.motionId === currentMotion?.id)?.answer;
	};

	if (isError) {
		return <div>Error loading questions...</div>; // Simplified error display
	}

	if ((isLoading && motions.length === 0) || (isFetching && !currentMotion)) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<div className="text-center">
					<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
					<p className="text-gray-600">Stellingen laden...</p>
				</div>
			</div>
		);
	}

	if (motions.length === 0 && !isLoading && !isFetching) {
		return <div>No motions found for the selected filters.</div>; // Simplified display
	}

	return (
		<div className="min-h-screen bg-gray-50 pb-48">
			{/* --- Sticky Header --- */}
			<div className="sticky top-0 bg-white shadow-sm z-20">
				<div className="container mx-auto px-4 py-3">
					<div className="flex items-center justify-between">
						<Link
							to="/compass/settings"
							className="flex items-center text-sm text-gray-600 hover:text-gray-900 whitespace-nowrap"
						>
							<ArrowLeft className="h-4 w-4 mr-2" />
							Instellingen
						</Link>
						<div className="max-w-lg flex-grow">
							<Progress value={progress} />
						</div>
						<div className="flex items-center gap-2 whitespace-nowrap">
							<span className="text-sm font-medium text-gray-700">
								Vraag {state.currentIndex + 1}
							</span>
							<Button
								variant="ghost"
								onClick={handleReset}
								className="text-gray-500 hover:text-gray-700"
								aria-label="Reset compass"
							>
								<RotateCcw className="h-4 w-4" />
							</Button>
						</div>
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
							<CardTitle className="text-2xl leading-tight">
								{currentMotion?.title}
							</CardTitle>
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
										<ul className="space-y-1.5">
											{displayedBulletPoints.map((point, index) => (
												<li
													key={`bullet-${currentMotion.id}-${index}`}
													className={`text-sm text-gray-800 flex items-start ${point.toLowerCase().trimStart().startsWith("verzoekt") ? "font-semibold" : ""}`}
												>
													<span className="text-primary-500 mr-2.5 mt-0.5">
														•
													</span>
													{point}
												</li>
											))}
										</ul>
										{currentMotion.bulletPoints.length >
											displayedBulletPoints.length && (
											<Button
												onClick={() =>
													setShowAllBulletPoints(!showAllBulletPoints)
												}
												className="p-0 h-auto mt-3 text-sm bg-transparent text-gray-800 hover:bg-transparent"
											>
												{showAllBulletPoints ? (
													<>
														<ChevronUp className="h-4 w-4 mr-1" />
														Toon alleen kernpunten
													</>
												) : (
													<>
														<ChevronDown className="h-4 w-4 mr-1" />
														Toon alle punten (
														{currentMotion.bulletPoints.length})
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
						<div className="grid grid-cols-3 gap-3 mb-4">
							<Button
								variant={
									getCurrentAnswer() === "agree" ? "primary" : "secondary"
								}
								size="lg"
								onClick={() => handleAnswer("agree")}
								className="h-20 text-lg font-semibold relative"
								disabled={submitAnswers.isPending}
							>
								{getCurrentAnswer() === "agree" && (
									<CheckCircle className="absolute top-2 right-2 h-5 w-5" />
								)}
								<div className="flex flex-col items-center">
									<ThumbsUp className="h-8 w-8 mb-1" />
									<span>Eens</span>
								</div>
							</Button>

							<Button
								variant={
									getCurrentAnswer() === "neutral" ? "primary" : "secondary"
								}
								size="lg"
								onClick={() => handleAnswer("neutral")}
								className="h-20 text-lg font-semibold relative"
								disabled={submitAnswers.isPending}
							>
								{getCurrentAnswer() === "neutral" && (
									<CheckCircle className="absolute top-2 right-2 h-5 w-5" />
								)}
								<div className="flex flex-col items-center">
									<Meh className="h-8 w-8 mb-1" />
									<span>Neutraal</span>
								</div>
							</Button>

							<Button
								variant={
									getCurrentAnswer() === "disagree" ? "primary" : "secondary"
								}
								size="lg"
								onClick={() => handleAnswer("disagree")}
								className="h-20 text-lg font-semibold relative"
								disabled={submitAnswers.isPending}
							>
								{getCurrentAnswer() === "disagree" && (
									<CheckCircle className="absolute top-2 right-2 h-5 w-5" />
								)}
								<div className="flex flex-col items-center">
									<ThumbsDown className="h-8 w-8 mb-1" />
									<span>Oneens</span>
								</div>
							</Button>
						</div>

						{state.answers.length >= 20 && (
							<div className="text-center mt-4">
								<Button
									onClick={() => handleSubmit(state.answers)}
									loading={submitAnswers.isPending}
									className="w-full md:w-auto"
									size="lg"
									variant={isLastQuestion ? "primary" : "secondary"}
								>
									{isLastQuestion
										? "🎉 Bekijk je eindresultaten"
										: "Bekijk je resultaten"}
									<span className="ml-2 font-normal opacity-80">
										({state.answers.length} antwoorden)
									</span>
									<ArrowRight className="h-4 w-4 ml-2" />
								</Button>
							</div>
						)}
					</div>
				</div>
			</div>
		</div>
	);
}
