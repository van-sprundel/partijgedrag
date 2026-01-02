import {
	Calendar,
	CheckCircle2,
	ChevronLeft,
	ClipboardList,
	Compass,
	Flame,
	Loader2,
	Save,
	Search,
	Settings,
	Users,
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
import { useCompassMotionsCount, useMotionCategories } from "../hooks/api";

interface CompassSettingsState {
	categoryIds: string[];
	coalitionOnly: boolean;
	search: string;
}

// Debounce hook
function useDebounce<T>(value: T, delay: number): T {
	const [debouncedValue, setDebouncedValue] = useState<T>(value);

	useEffect(() => {
		const handler = setTimeout(() => {
			setDebouncedValue(value);
		}, delay);

		return () => {
			clearTimeout(handler);
		};
	}, [value, delay]);

	return debouncedValue;
}

export function CompassSettingsPage() {
	const navigate = useNavigate();
	const [searchParams] = useSearchParams();
	const { data: availableCategories = [], isLoading } = useMotionCategories();

	// Parse current query parameters
	const initialSettings = useMemo(() => {
		const categoryIds =
			searchParams.get("categoryIds")?.split(",").filter(Boolean) || [];
		const coalitionOnly = searchParams.has("after")
			? searchParams.get("after") === "2024-07-02"
			: true;
		const search = searchParams.get("search") || "";

		return {
			categoryIds,
			coalitionOnly,
			search,
		};
	}, [searchParams]);

	const [settings, setSettings] =
		useState<CompassSettingsState>(initialSettings);

	const debouncedSettings = useDebounce(settings, 500);

	const { data: motionCount, isFetching: isCountFetching } =
		useCompassMotionsCount({
			categoryIds:
				debouncedSettings.categoryIds.length > 0
					? debouncedSettings.categoryIds
					: undefined,
			after: debouncedSettings.coalitionOnly ? new Date("2024-07-02") : undefined,
			search: debouncedSettings.search || undefined,
		});

	useEffect(() => {
		setSettings(initialSettings);
	}, [initialSettings]);

	const handleCategoryToggle = (categoryId: string) => {
		setSettings((prev) => ({
			...prev,
			categoryIds: prev.categoryIds.includes(categoryId)
				? prev.categoryIds.filter((id) => id !== categoryId)
				: [...prev.categoryIds, categoryId],
		}));
	};

	const handleCoalitionToggle = () => {
		setSettings((prev) => ({
			...prev,
			coalitionOnly: !prev.coalitionOnly,
		}));
	};

	const handleSelectNone = () => {
		setSettings((prev) => ({
			...prev,
			categoryIds: [],
		}));
	};

	const handleStartCompass = () => {
		const params = new URLSearchParams();

		if (settings.categoryIds.length > 0) {
			params.set("categoryIds", settings.categoryIds.join(","));
		}

		if (settings.coalitionOnly) {
			params.set("after", "2024-07-02");
		}

		if (settings.search) {
			params.set("search", settings.search);
		}

		const queryString = params.toString();
		navigate(`/compass${queryString ? `?${queryString}` : ""}`);
	};

	const isEditingExistingFilters =
		searchParams.has("categoryIds") || searchParams.has("after") || searchParams.has("search");

	if (isLoading) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center pt-16">
				<div className="text-center">
					<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
					<p className="text-gray-600">Instellingen laden...</p>
				</div>
			</div>
		);
	}

	return (
		<div className="bg-gray-50">
			<div className="bg-white border-b border-gray-200 sticky top-0 z-10">
				<div className="container mx-auto px-4 py-6 max-w-4xl">
					<div className="flex items-center justify-between">
						<div className="flex items-center">
							<Settings className="h-8 w-8 text-primary-600 mr-3" />
							<div>
								<h1 className="text-3xl font-bold text-gray-900">
									Stemwijzer Instellingen
								</h1>
								<p className="text-gray-600">
									Stel je voorkeuren in voor de politieke stellingen
								</p>
								<div className="text-sm text-gray-500 mt-2 flex items-center gap-2">
									<span className="font-semibold text-blue-600">
										{isCountFetching ? (
											<Loader2 className="h-4 w-4 animate-spin" />
										) : (
											<>{motionCount?.count ?? 0} stellingen beschikbaar</>
										)}
									</span>
									<span className="text-xs">
										(je beantwoordt er minimaal 20)
									</span>
								</div>
							</div>
						</div>
						<Link
							to="/"
							className="text-gray-600 hover:text-gray-900 flex items-center"
						>
							<ChevronLeft className="h-4 w-4 mr-1" />
							Terug naar home
						</Link>
					</div>
				</div>
			</div>

			<div className="container mx-auto px-4 py-8">
				<div className="max-w-4xl mx-auto">
					<div className="grid gap-6">
						{/* Coalition Period Setting */}
						<Card>
							<CardHeader>
								<div className="flex items-center">
									<Calendar className="h-6 w-6 text-primary-600 mr-3" />
									<div>
										<CardTitle>Periode</CardTitle>
										<CardDescription>
											Kies welke stellingen je wilt zien op basis van de
											regeerperiode
										</CardDescription>
									</div>
								</div>
							</CardHeader>
							<CardContent>
								<div className="space-y-4">
									<button
										type="button"
										className={`w-full border-2 rounded-lg p-4 cursor-pointer transition-all text-left ${
											settings.coalitionOnly
												? "border-primary-500 bg-primary-50"
												: "border-gray-200 hover:border-gray-300"
										}`}
										onClick={handleCoalitionToggle}
									>
										<div className="flex items-center justify-between">
											<div className="flex items-center">
												<Users className="h-5 w-5 text-primary-600 mr-3" />
												<div>
													<h3 className="font-semibold text-gray-900">
														Alleen huidige coalitie (aanbevolen)
													</h3>
													<p className="text-sm text-gray-600">
														Stellingen vanaf 2 juli 2024 - periode van het
														huidige kabinet (Schoof I)
													</p>
												</div>
											</div>
											{settings.coalitionOnly && (
												<CheckCircle2 className="h-5 w-5 text-primary-600" />
											)}
										</div>
									</button>

									<button
										type="button"
										className={`w-full border-2 rounded-lg p-4 cursor-pointer transition-all text-left ${
											!settings.coalitionOnly
												? "border-primary-500 bg-primary-50"
												: "border-gray-200 hover:border-gray-300"
										}`}
										onClick={handleCoalitionToggle}
									>
										<div className="flex items-center justify-between">
											<div className="flex items-center">
												<Calendar className="h-5 w-5 text-gray-600 mr-3" />
												<div>
													<h3 className="font-semibold text-gray-900">
														Alle beschikbare stellingen
													</h3>
													<p className="text-sm text-gray-600">
														Inclusief stellingen van vorige kabinetten (vanaf
														2023)
													</p>
												</div>
											</div>
											{!settings.coalitionOnly && (
												<CheckCircle2 className="h-5 w-5 text-primary-600" />
											)}
										</div>
									</button>
								</div>
							</CardContent>
						</Card>

						{/* Category Selection */}
						<Card>
							<CardHeader>
								<div className="flex items-center justify-between">
									<div className="flex items-center">
										<CheckCircle2 className="h-6 w-6 text-primary-600 mr-3" />
										<div>
											<CardTitle>Onderwerpen</CardTitle>
											<CardDescription>
												Selecteer specifieke onderwerpen of laat leeg voor alle
												onderwerpen
											</CardDescription>
										</div>
									</div>
									<div className="flex gap-2">
										<Button
											variant="ghost"
											size="sm"
											onClick={handleSelectNone}
											disabled={settings.categoryIds.length === 0}
										>
											Alles deselecteren
										</Button>
									</div>
								</div>
							</CardHeader>
							<CardContent>
								{settings.categoryIds.length === 0 && (
									<div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
										<p className="text-blue-800 text-sm">
											<strong>Alle onderwerpen:</strong> Er zijn geen specifieke
											onderwerpen geselecteerd. Je krijgt een willekeurige
											selectie van alle beschikbare stellingen.
										</p>
									</div>
								)}

								{/* Hot Topics Section */}
								{availableCategories.filter((cat) => cat.type === "hot_topic")
									.length > 0 && (
									<div className="mb-6">
										<h3 className="text-lg font-semibold text-gray-900 mb-3 flex items-center">
											<Flame className="h-5 w-5 mr-2 text-red-500" /> Hot Topics
											<span className="ml-2 text-sm font-normal text-gray-600">
												- actuele onderwerpen
											</span>
										</h3>
										<div className="grid grid-cols-1 md:grid-cols-2 gap-3">
											{availableCategories
												.filter((cat) => cat.type === "hot_topic")
												.map((category) => (
													<button
														type="button"
														key={category.id}
														className={`w-full border-2 rounded-lg p-3 cursor-pointer transition-all text-left ${
															settings.categoryIds.includes(category.id)
																? "border-red-500 bg-red-50"
																: "border-gray-200 hover:border-red-300"
														}`}
														onClick={() => handleCategoryToggle(category.id)}
													>
														<div className="flex items-center justify-between">
															<span className="font-medium text-gray-900">
																{category.name}
															</span>
															{settings.categoryIds.includes(category.id) && (
																<CheckCircle2 className="h-4 w-4 text-red-600" />
															)}
														</div>
													</button>
												))}
										</div>
									</div>
								)}

								{/* Regular Categories */}
								<div>
									<h3 className="text-lg font-semibold text-gray-900 mb-3 flex items-center">
										<ClipboardList className="h-5 w-5 mr-2 text-blue-500" />{" "}
										Beleidsagenda
									</h3>

									<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
										{availableCategories
											.filter((cat) => cat.type === "generic")
											.map((category) => (
												<button
													type="button"
													key={category.id}
													className={`w-full border-2 rounded-lg p-3 cursor-pointer transition-all text-left ${
														settings.categoryIds.includes(category.id)
															? "border-primary-500 bg-primary-50"
															: "border-gray-200 hover:border-gray-300"
													}`}
													onClick={() => handleCategoryToggle(category.id)}
												>
													<div className="flex items-center justify-between">
														<span className="font-medium text-gray-900">
															{category.name}
														</span>
														{settings.categoryIds.includes(category.id) && (
															<CheckCircle2 className="h-4 w-4 text-primary-600" />
														)}
													</div>
												</button>
											))}
									</div>
								</div>
							</CardContent>
						</Card>

						{/* Free Text Search */}
						<Card>
							<CardHeader>
								<div className="flex items-center">
									<Search className="h-6 w-6 text-primary-600 mr-3" />
									<div>
										<CardTitle>Vrije tekst</CardTitle>
										<CardDescription>
											Zoek naar specifieke woorden of begrippen in alle
											stellingen
										</CardDescription>
									</div>
								</div>
							</CardHeader>
							<CardContent>
								<div className="space-y-3">
									<input
										type="text"
										placeholder="Bijv. klimaat, energietransitie, onderwijs..."
										value={settings.search}
										onChange={(e) =>
											setSettings((prev) => ({
												...prev,
												search: e.target.value,
											}))
										}
										className="w-full border-2 border-gray-200 rounded-lg p-3 focus:outline-none focus:border-primary-500 transition-colors"
									/>
									<p className="text-sm text-gray-600">
										{settings.search
											? `Zoeken naar: "${settings.search}"`
											: "Laat leeg om op alle stellingen te zoeken"}
									</p>
								</div>
							</CardContent>
						</Card>

						{/* Start Button */}
						<div className="flex justify-center pt-6">
							<Button
								onClick={handleStartCompass}
								size="lg"
								className="flex items-center px-12 py-4 text-lg"
								disabled={isCountFetching || motionCount?.count === 0}
							>
								{isEditingExistingFilters ? (
									<>
										<Save className="h-5 w-5 mr-2" />
										Filters toepassen
									</>
								) : (
									<>
										<Compass className="h-5 w-5 mr-2" />
										Start Stemwijzer
									</>
								)}
							</Button>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}
