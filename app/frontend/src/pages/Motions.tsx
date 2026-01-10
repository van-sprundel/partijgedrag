import {
	Calendar,
	ChevronLeft,
	ChevronRight,
	ExternalLink,
	FileText,
	Filter,
	Search,
	X,
} from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../components/ui/Card";
import { useMotions, useMotionVotes, useParties } from "../hooks/api";
import { formatDate } from "../lib/utils";

const ITEMS_PER_PAGE = 20;

// Component to display voting results for a motion
const MotionVoteResult = ({ motionId }: { motionId: string }) => {
	const { data: voteData, isLoading } = useMotionVotes(motionId);

	if (isLoading) {
		return <div className="text-xs text-gray-400">Laden...</div>;
	}

	if (!voteData?.partyPositions || voteData.partyPositions.length === 0) {
		return <div className="text-xs text-gray-500">Geen stemdata</div>;
	}

	const voteCounts = voteData.partyPositions.reduce(
		(acc, p) => {
			acc[p.position] = (acc[p.position] || 0) + p.count;
			return acc;
		},
		{} as Record<string, number>,
	);

	const totalVotes = Object.values(voteCounts).reduce((a, b) => a + b, 0);

	return (
		<div className="flex items-center gap-2">
			<div className="flex items-center gap-1 text-xs">
				<div className="w-2 h-2 bg-green-500 rounded-full"></div>
				<span>{voteCounts.FOR || 0}</span>
			</div>
			<div className="flex items-center gap-1 text-xs">
				<div className="w-2 h-2 bg-red-500 rounded-full"></div>
				<span>{voteCounts.AGAINST || 0}</span>
			</div>
			<div className="flex items-center gap-1 text-xs">
				<div className="w-2 h-2 bg-yellow-500 rounded-full"></div>
				<span>{voteCounts.NEUTRAL || 0}</span>
			</div>
			<span className="text-xs text-gray-400">({totalVotes})</span>
		</div>
	);
};

export function MotionsPage() {
	const [currentPage, setCurrentPage] = useState(1);
	const [showOnlyWithVotes, setShowOnlyWithVotes] = useState(false);
	const [searchInput, setSearchInput] = useState("");
	const [searchQuery, setSearchQuery] = useState("");
	const [selectedPartyIds, setSelectedPartyIds] = useState<string[]>([]);
	const offset = (currentPage - 1) * ITEMS_PER_PAGE;

	const { data: parties } = useParties({ activeOnly: true });

	// Debounce search input
	useEffect(() => {
		const timer = setTimeout(() => {
			setSearchQuery(searchInput);
			setCurrentPage(1); // Reset to first page when search changes
		}, 500);

		return () => clearTimeout(timer);
	}, [searchInput]);

	const { data, isLoading, error } = useMotions({
		limit: ITEMS_PER_PAGE,
		offset,
		withVotes: showOnlyWithVotes,
		search: searchQuery || undefined,
		partyIds: selectedPartyIds.length > 0 ? selectedPartyIds : undefined,
	});

	const totalPages = data ? Math.ceil(data.total / ITEMS_PER_PAGE) : 0;

	if (error) {
		return (
			<div className="min-h-screen bg-gray-50 flex items-center justify-center">
				<Card className="max-w-md mx-auto">
					<CardHeader>
						<CardTitle>Fout bij laden</CardTitle>
						<CardDescription>
							Er ging iets mis bij het laden van de moties.
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

	const handlePageChange = (page: number) => {
		setCurrentPage(page);
		window.scrollTo({ top: 0, behavior: "smooth" });
	};

	const handleFilterChange = () => {
		setShowOnlyWithVotes(!showOnlyWithVotes);
		setCurrentPage(1); // Reset to first page when filter changes
		window.scrollTo({ top: 0, behavior: "smooth" });
	};

	const getDescription = () => {
		const baseText = showOnlyWithVotes
			? "Moties uit de Tweede Kamer met stemresultaten"
			: "Alle moties uit de Tweede Kamer";
		const searchText = searchQuery ? ` - Zoekresultaten voor "${searchQuery}"` : "";
		const partyFilterText =
			selectedPartyIds.length > 0
				? ` - Gefilterd op ${selectedPartyIds.length} ${selectedPartyIds.length === 1 ? "partij" : "partijen"}`
				: "";
		const count = isLoading ? "..." : `${data?.total || 0} moties`;
		return `${baseText}${searchText}${partyFilterText} (${count})`;
	};

	const handleClearSearch = () => {
		setSearchInput("");
		setSearchQuery("");
	};

	return (
		<div className="bg-gray-50">
			<div className="container mx-auto px-4 py-8 max-w-7xl">
				<Card>
					<CardHeader>
						<div className="flex flex-col gap-4">
							<div className="flex items-center justify-between">
								<div>
									<CardTitle>Moties</CardTitle>
									<CardDescription>{getDescription()}</CardDescription>
								</div>
								<div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg border">
									<Filter className="h-4 w-4 text-gray-600" />
									<label className="flex items-center gap-2 text-sm font-medium text-gray-700 cursor-pointer">
										<input
											type="checkbox"
											checked={showOnlyWithVotes}
											onChange={handleFilterChange}
											disabled={isLoading}
											className="rounded border-gray-300 text-primary-600 focus:ring-primary-500 focus:ring-offset-0 focus:ring-2 disabled:opacity-50"
										/>
										Alleen met stemresultaten
									</label>
									{isLoading && (
										<div className="animate-spin h-4 w-4 border-2 border-primary-600 border-t-transparent rounded-full"></div>
									)}
								</div>
							</div>

							{/* Search Bar */}
							<div className="relative">
								<div className="flex items-center">
									<Search className="absolute left-3 h-5 w-5 text-gray-400" />
									<input
										type="text"
										value={searchInput}
										onChange={(e) => setSearchInput(e.target.value)}
										placeholder="Zoek in moties... (bijv. 'chroomverf', 'CO2-heffing', '2015Z20433')"
										className="w-full pl-10 pr-10 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
									/>
									{searchInput && (
										<button
											type="button"
											onClick={handleClearSearch}
											className="absolute right-3 p-1 hover:bg-gray-100 rounded-full transition-colors"
											aria-label="Wis zoekopdracht"
										>
											<X className="h-4 w-4 text-gray-400" />
										</button>
									)}
								</div>
								{searchQuery && (
									<p className="mt-2 text-sm text-gray-600">
										Zoeken naar: <span className="font-medium">"{searchQuery}"</span>
									</p>
								)}
							</div>

							{/* Party Filter */}
							{parties && parties.length > 0 && (
								<div className="p-4 bg-gray-50 rounded-lg border">
									<div className="flex items-center gap-2 mb-3">
										<Filter className="h-4 w-4 text-gray-600" />
										<h3 className="text-sm font-medium text-gray-700">
											Filter op partijen
										</h3>
										{selectedPartyIds.length > 0 && (
											<button
												type="button"
												onClick={() => {
													setSelectedPartyIds([]);
													setCurrentPage(1);
												}}
												className="ml-auto text-xs text-primary-600 hover:text-primary-700 font-medium"
											>
												Wis filter ({selectedPartyIds.length})
											</button>
										)}
									</div>
									<p className="text-xs text-gray-600 mb-3">
										Selecteer partijen om alleen moties te tonen waar deze partijen voor
										of tegen hebben gestemd
									</p>
									<div className="flex flex-wrap gap-2">
										{parties.map((party) => {
											const isSelected = selectedPartyIds.includes(party.id);
											return (
												<button
													key={party.id}
													type="button"
													onClick={() => {
														setSelectedPartyIds((prev) =>
															prev.includes(party.id)
																? prev.filter((id) => id !== party.id)
																: [...prev, party.id],
														);
														setCurrentPage(1);
													}}
													className={`px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
														isSelected
															? "bg-primary-600 text-white"
															: "bg-white text-gray-700 border border-gray-300 hover:border-primary-500"
													}`}
												>
													{party.shortName}
												</button>
											);
										})}
									</div>
								</div>
							)}
						</div>
					</CardHeader>

					<CardContent className="p-0">
						{/* Table Header */}
						<div className="hidden md:grid md:grid-cols-12 gap-4 px-6 py-3 bg-gray-50 border-b font-medium text-sm text-gray-700">
							<div className="col-span-5">Motie</div>
							<div className="col-span-2">Datum</div>
							<div className="col-span-3">Stemresultaat</div>
							<div className="col-span-2">Acties</div>
						</div>

						{/* Table Rows */}
						<div className="divide-y divide-gray-200">
							{isLoading ? (
								// Skeleton loading state
								<div className="animate-pulse space-y-0">
									{Array.from({ length: ITEMS_PER_PAGE }, (_, i) => (
										<div
											key={`skeleton-${i}`}
											className="grid grid-cols-1 md:grid-cols-12 gap-4 px-6 py-4"
										>
											<div className="md:col-span-5 space-y-2">
												<div className="h-5 bg-gray-200 rounded w-3/4"></div>
												<div className="h-4 bg-gray-200 rounded w-1/2"></div>
												<div className="h-3 bg-gray-200 rounded w-1/4"></div>
											</div>
											<div className="md:col-span-2">
												<div className="h-4 bg-gray-200 rounded w-24"></div>
											</div>
											<div className="md:col-span-3">
												<div className="h-4 bg-gray-200 rounded w-32"></div>
											</div>
											<div className="md:col-span-2 flex gap-2">
												<div className="h-8 bg-gray-200 rounded w-16"></div>
												<div className="h-8 bg-gray-200 rounded w-8"></div>
											</div>
										</div>
									))}
								</div>
							) : !data || data.motions.length === 0 ? (
								// Empty state
								<div className="px-6 py-12 text-center text-gray-500">
									{searchQuery
										? `Geen moties gevonden voor "${searchQuery}"`
										: "Geen moties beschikbaar"}
								</div>
							) : (
								// Actual data
								data.motions.map((motion) => (
								<div
									key={motion.id}
									className="grid grid-cols-1 md:grid-cols-12 gap-4 px-6 py-4 hover:bg-gray-50 transition-colors"
								>
									{/* Motion Info */}
									<div className="md:col-span-5">
										<Link
											to={`/motion/${motion.id}`}
											className="block hover:text-primary-600 transition-colors"
										>
											<h3 className="font-semibold text-gray-900 mb-1 line-clamp-2">
												{motion.title}
											</h3>
											{motion.shortTitle &&
												motion.shortTitle !== motion.title && (
													<p className="text-sm text-gray-600 mb-2 line-clamp-1">
														{motion.shortTitle}
													</p>
												)}
											<div className="flex items-center gap-3 text-xs text-gray-500">
												{motion.motionNumber && (
													<span className="flex items-center">
														<FileText className="h-3 w-3 mr-1" />
														{motion.motionNumber}
													</span>
												)}
												<span className="px-2 py-1 bg-gray-100 rounded-full">
													{motion.status}
												</span>
											</div>
										</Link>
									</div>

									{/* Date */}
									<div className="md:col-span-2 flex items-center">
										<div className="flex items-center text-sm text-gray-600">
											<Calendar className="h-4 w-4 mr-2" />
											<span>
												{motion.createdAt
													? formatDate(motion.createdAt)
													: "Onbekend"}
											</span>
										</div>
									</div>

									{/* Voting Results */}
									<div className="md:col-span-3 flex items-center">
										<MotionVoteResult motionId={motion.id} />
									</div>

									{/* Actions */}
									<div className="md:col-span-2 flex items-center gap-2">
										<Link
											to={`/motion/${motion.id}`}
											className="px-3 py-1 text-sm bg-primary-600 text-white rounded-md hover:bg-primary-700 transition-colors"
										>
											Details
										</Link>
										{motion.motionNumber && motion.did && (
											<a
												href={`https://www.tweedekamer.nl/kamerstukken/moties/detail?id=${motion.motionNumber}&did=${motion.did}`}
												target="_blank"
												rel="noopener noreferrer"
												className="px-2 py-1 text-sm bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 transition-colors"
											>
												<ExternalLink className="h-3 w-3" />
											</a>
										)}
									</div>
								</div>
								))
							)}
						</div>

						{/* Pagination */}
						{!isLoading && totalPages > 1 && (
							<div className="flex items-center justify-between px-6 py-4 border-t bg-gray-50">
								<div className="text-sm text-gray-600">
									Pagina {currentPage} van {totalPages} ({data?.total} moties)
								</div>

								<div className="flex items-center space-x-2">
									<button
										type="button"
										onClick={() =>
											handlePageChange(Math.max(1, currentPage - 1))
										}
										disabled={currentPage <= 1}
										className="flex items-center px-3 py-1 text-sm bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
									>
										<ChevronLeft className="h-4 w-4 mr-1" />
										Vorige
									</button>

									{/* Page Numbers */}
									<div className="flex items-center space-x-1">
										{Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
											let pageNumber: number;
											if (totalPages <= 5) {
												pageNumber = i + 1;
											} else if (currentPage <= 3) {
												pageNumber = i + 1;
											} else if (currentPage >= totalPages - 2) {
												pageNumber = totalPages - 4 + i;
											} else {
												pageNumber = currentPage - 2 + i;
											}

											return (
												<button
													type="button"
													key={pageNumber}
													onClick={() => handlePageChange(pageNumber)}
													className={`px-3 py-1 text-sm rounded-md transition-colors ${
														currentPage === pageNumber
															? "bg-primary-600 text-white"
															: "bg-white border border-gray-300 hover:bg-gray-50"
													}`}
												>
													{pageNumber}
												</button>
											);
										})}
									</div>

									<button
										type="button"
										onClick={() =>
											handlePageChange(Math.min(totalPages, currentPage + 1))
										}
										disabled={currentPage >= totalPages}
										className="flex items-center px-3 py-1 text-sm bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
									>
										Volgende
										<ChevronRight className="h-4 w-4 ml-1" />
									</button>
								</div>
							</div>
						)}
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
