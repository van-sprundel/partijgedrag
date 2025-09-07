import { ArrowLeft, ArrowRight, CheckCircle, RotateCcw } from "lucide-react";
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
import type { UserAnswer } from "../lib/api";
import { calculateProgress, truncateText } from "../lib/utils";

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

	const {
		data: motions = [],
		isLoading,
		isError,
		error,
	} = useCompassMotions(20, [], filterParams.categoryIds, filterParams.after);

	const submitAnswers = useSubmitAnswers();

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
	const isLastQuestion = state.currentIndex === totalMotions - 1;

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
			if (!isLastQuestion) {
				setState((prev) => ({
					...prev,
					currentIndex: prev.currentIndex + 1,
				}));
			}
			// Don't auto-submit on last question - let user control when to submit
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
		if (sessionId) {
			// Navigate to new compass session
			navigate("/compass");
		}
	};

	const getCurrentAnswer = (): Answer | undefined => {
		return state.answers.find((a) => a.motionId === currentMotion?.id)?.answer;
	};

	if (isError) {
		console.log(error);
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<Card className="max-w-md mx-auto">
					<CardHeader>
						<CardTitle>Fout bij laden</CardTitle>
						<CardDescription>
							Er is een fout opgetreden bij het laden van de stellingen.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="flex gap-2">
							<Button
								onClick={() => window.location.reload()}
								variant="primary"
							>
								Probeer opnieuw
							</Button>
							<Link to="/">
								<Button variant="secondary">Terug naar home</Button>
							</Link>
						</div>
					</CardContent>
				</Card>
			</div>
		);
	}

	if (isLoading) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<div className="text-center">
					<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
					<p className="text-gray-600">Stellingen laden...</p>
				</div>
			</div>
		);
	}

	if (motions.length === 0) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<Card className="max-w-md mx-auto">
					<CardHeader>
						<CardTitle>Geen stellingen gevonden</CardTitle>
						<CardDescription>
							Er zijn geen stellingen gevonden die overeenkomen met de
							geselecteerde filters.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<Link to={`/compass/settings?${searchParams.toString()}`}>
							<Button variant="primary">Pas filters aan</Button>
						</Link>
					</CardContent>
				</Card>
			</div>
		);
	}

	// Show completion message if continuing session and no more unanswered motions
	if (sessionId && unansweredMotions.length === 0 && state.answers.length > 0) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<Card className="max-w-md mx-auto">
					<CardHeader>
						<CardTitle>🎉 Alle stellingen beantwoord!</CardTitle>
						<CardDescription>
							Je hebt alle beschikbare stellingen beantwoord. Bekijk je
							resultaten of start een nieuwe stemwijzer.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="flex gap-2">
							<Link to={`/results/${sessionId}`}>
								<Button variant="primary">Bekijk resultaten</Button>
							</Link>
							<Link to="/compass">
								<Button variant="secondary">Nieuwe stemwijzer</Button>
							</Link>
						</div>
					</CardContent>
				</Card>
			</div>
		);
	}

	return (
		<div className="min-h-screen bg-gray-50 pb-32">
			{/* Header */}
			<div className="bg-white shadow-sm">
				<div className="container mx-auto px-4 py-4">
					<div className="flex items-center justify-between">
						<Link
							to="/compass/settings"
							className="flex items-center text-gray-600 hover:text-gray-900"
						>
							<ArrowLeft className="h-5 w-5 mr-2" />
							Terug naar instellingen
						</Link>
						<div className="flex items-center gap-4">
							{sessionId && (
								<Link to={`/results/${sessionId}`}>
									<Button variant="ghost" size="sm">
										Huidige resultaten
									</Button>
								</Link>
							)}
							<span className="text-sm text-gray-600">
								{sessionId
									? `${state.currentIndex + 1} van ${unansweredMotions.length} nieuwe stellingen`
									: `Vraag ${state.currentIndex + 1} van ${motions.length}`}
							</span>
							<Button
								variant="ghost"
								size="sm"
								onClick={handleReset}
								className="text-gray-500 hover:text-gray-700"
							>
								<RotateCcw className="h-4 w-4 mr-1" />
								{sessionId ? "Nieuwe sessie" : "Opnieuw"}
							</Button>
						</div>
					</div>
				</div>
			</div>

			{/* Progress */}
			<div className="container mx-auto px-4 py-6">
				{/* Filter info */}
				{(filterParams.categoryIds || filterParams.after) && (
					<div className="max-w-4xl mx-auto mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
						<div className="flex items-start justify-between">
							<div>
								<p className="text-green-800 font-medium mb-1">
									🎯 Actieve filters:
								</p>
								<ul className="text-green-700 text-sm space-y-1">
									{filterParams.categoryIds && (
										<li>
											• <strong>Onderwerpen:</strong>{" "}
											{filterParams.categoryIds.length} categorie(ën)
											geselecteerd
										</li>
									)}
									{filterParams.after && (
										<li>
											• <strong>Periode:</strong> Vanaf{" "}
											{filterParams.after.toLocaleDateString("nl-NL")} (huidige
											coalitie)
										</li>
									)}
								</ul>
							</div>
							<Link
								to={`/compass/settings?${new URLSearchParams(
									Object.entries(filterParams)
										.filter(([, value]) => value !== undefined)
										.map(([key, value]) => [key, String(value)]),
								).toString()}`}
							>
								<Button
									variant="ghost"
									size="sm"
									className="text-green-700 hover:text-green-800"
								>
									Wijzig filters
								</Button>
							</Link>
						</div>
					</div>
				)}

				{sessionId && state.answers.length > 0 && (
					<div className="max-w-4xl mx-auto mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
						<p className="text-blue-800">
							Je zet een bestaande sessie voort met {state.answers.length}{" "}
							eerder beantwoorde stellingen.
							{unansweredMotions.length > 0 &&
								` We tonen je nu ${unansweredMotions.length} nieuwe stellingen.`}
						</p>
					</div>
				)}
				<Progress value={progress} showValue className="mb-8" />

				{/* Question Card */}
				<div className="max-w-4xl mx-auto mb-8">
					<Card
						className={`mb-6 ${getCurrentAnswer() ? "ring-2 ring-green-200 bg-green-50" : ""}`}
					>
						<CardHeader>
							<div className="flex items-start justify-between">
								<div className="flex-1">
									<div className="flex items-center gap-3 mb-3">
										<CardTitle className="text-2xl">
											{currentMotion?.title}
										</CardTitle>
										{getCurrentAnswer() && (
											<span className="text-green-600 text-xl">
												{getCurrentAnswer() === "agree"
													? "👍"
													: getCurrentAnswer() === "neutral"
														? "🤷"
														: "👎"}
											</span>
										)}
									</div>
									{getCurrentAnswer() && (
										<div className="inline-flex items-center gap-2 mb-2 text-sm font-medium text-green-700">
											<span>✓</span>
											<span>
												Beantwoord:{" "}
												{getCurrentAnswer() === "agree"
													? "Eens"
													: getCurrentAnswer() === "neutral"
														? "Neutraal"
														: "Oneens"}
											</span>
										</div>
									)}
									{currentMotion?.description && (
										<CardDescription className="text-base leading-relaxed">
											{state.showExplanation
												? currentMotion.description
												: truncateText(currentMotion.description, 150)}
											{currentMotion.description.length > 150 && (
												<button
													type="button"
													onClick={() =>
														setState((prev) => ({
															...prev,
															showExplanation: !prev.showExplanation,
														}))
													}
													className="ml-2 text-primary-600 hover:text-primary-700 font-medium"
												>
													{state.showExplanation ? "Minder" : "Meer"}
												</button>
											)}
										</CardDescription>
									)}
								</div>
								{currentMotion?.category && (
									<span className="bg-primary-100 text-primary-800 text-xs font-medium px-2.5 py-0.5 rounded-full ml-4">
										{currentMotion.category}
									</span>
								)}
							</div>
						</CardHeader>

						{/* Bullet Points */}
						{currentMotion?.bulletPoints?.length > 0 && (
							<CardContent className="pt-0">
								<div className="bg-gray-50 rounded-lg p-4">
									<h4 className="font-medium text-gray-900 mb-2">
										Kernpunten:
									</h4>
									<ul className="space-y-1">
										{currentMotion.bulletPoints.map((point, index) => (
											<li
												key={`bullet-${currentMotion.id}-${index}`}
												className={`text-sm text-gray-700 flex items-start ${point.toLowerCase().trimStart().startsWith("verzoekt") ? "font-bold" : ""}`}
											>
												<span className="text-primary-500 mr-2 mt-0.5">•</span>
												{point}
											</li>
										))}
									</ul>
								</div>
							</CardContent>
						)}
					</Card>

					{/* Navigation */}
					<div className="flex justify-between items-center">
						<span className="text-sm text-gray-500">
							{sessionId ? (
								<>
									{state.currentIndex + 1} van {unansweredMotions.length} nieuwe
									<span className="ml-2 text-xs text-blue-600">
										({state.answers.length} eerder beantwoord)
									</span>
								</>
							) : (
								<>
									{state.answers.length} van {motions.length} beantwoord
								</>
							)}
						</span>
					</div>
				</div>
			</div>

			{/* Sticky Answer Buttons */}
			<div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 shadow-lg z-10">
				<div className="container mx-auto px-4 py-4">
					<div className="max-w-4xl mx-auto">
						{/* Answer Buttons */}
						<div className="grid grid-cols-3 gap-4 mb-4">
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
									<CheckCircle className="absolute top-3 right-3 h-5 w-5" />
								)}
								<div className="flex flex-col items-center">
									<span className="text-2xl mb-1">👍</span>
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
									<CheckCircle className="absolute top-3 right-3 h-5 w-5" />
								)}
								<div className="flex flex-col items-center">
									<span className="text-2xl mb-1">🤷</span>
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
									<CheckCircle className="absolute top-3 right-3 h-5 w-5" />
								)}
								<div className="flex flex-col items-center">
									<span className="text-2xl mb-1">👎</span>
									<span>Oneens</span>
								</div>
							</Button>
						</div>

						{/* Submit Button */}
						{state.answers.length >= 5 && (
							<div className="text-center">
								<Button
									onClick={() => handleSubmit(state.answers)}
									loading={submitAnswers.isPending}
									className="flex items-center mx-auto"
									size="lg"
									variant={isLastQuestion ? "primary" : "secondary"}
								>
									{isLastQuestion
										? `🎉 Bekijk je resultaten (${state.answers.length} antwoorden)`
										: `Bekijk resultaten (${state.answers.length} antwoorden)`}
									<ArrowRight className="h-4 w-4 ml-2" />
								</Button>
								{isLastQuestion && (
									<p className="text-sm text-gray-600 mt-2">
										Je hebt alle stellingen beantwoord! Klik om je resultaten te
										bekijken.
									</p>
								)}
							</div>
						)}
					</div>
				</div>
			</div>
		</div>
	);
}
