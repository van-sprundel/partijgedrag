import { ArrowRight, Clock, Search, Users, Vote } from "lucide-react";
import { Link } from "react-router-dom";

import { Card, CardContent } from "@/components/ui/Card";
import { useGetMotionStatistics } from "@/hooks/useGetMotionStatistics";
import { useRecentSessions } from "@/hooks/api";
import logo from "../assets/tug.png";
import { Button } from "../components/ui/Button";
import { formatPercentage } from "../lib/utils";

export function HomePage() {
	const { data: motionStatisticsData, isLoading } = useGetMotionStatistics();
	const { sessions, isLoading: isLoadingSessions } = useRecentSessions();

	return (
		<div className="pb-16 ">
			{/* Hero Section */}
			<div className="bg-gradient-to-b from-blue-50 to-white py-16 ">
				<div className="container mx-auto px-4 text-center">
					<div className="flex justify-center mb-6">
						<div className="flex items-center space-x-2">
							<img src={logo} alt="Partijgedrag Logo" className="size-24" />
						</div>
					</div>

					<h1 className="text-5xl font-bold text-gray-900 mb-4">
						Partijgedrag
					</h1>
					<p className="text-xl text-gray-600 mb-2">
						Geen woorden maar daden: stemgedrag in de Tweede Kamer
					</p>
					<p className="text-lg text-gray-500 mb-8">
						Laat u zien of politieke partijen met hun stemgedrag úw idealen
						hebben behartigd.
					</p>

					<div className="flex flex-col sm:flex-row gap-4 justify-center">
						<Link to="/compass/settings">
							<Button
								size="lg"
								className="text-lg px-8 py-4 bg-blue-600 hover:bg-blue-700"
							>
								<Vote className="mr-2 h-5 w-5" />
								Start Kieswijzer
								<ArrowRight className="ml-2 h-5 w-5" />
							</Button>
						</Link>
						<Link to="/motions">
							<Button
								variant="secondary"
								size="lg"
								className="text-lg px-8 py-4"
							>
								<Search className="mr-2 h-5 w-5" />
								Bekijk Moties
							</Button>
						</Link>
					</div>
				</div>
			</div>

			{/* Recent Sessions Section */}
			{!isLoadingSessions && sessions.length > 0 && (
				<div className="bg-white border-b border-gray-200">
					<div className="container mx-auto px-4 py-8">
						<div className="mb-4 flex items-center justify-between">
							<div>
								<h2 className="text-2xl font-bold text-gray-900">
									Eerdere resultaten
								</h2>
								<p className="text-sm text-gray-600 mt-1">
									Ga terug naar je vorige kieswijzer sessies
								</p>
							</div>
						</div>
						<div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
							{sessions.slice(0, 3).map((session) => {
								const topParty = session.data?.partyResults[0];
								const formattedDate = new Date(
									session.createdAt,
								).toLocaleDateString("nl-NL", {
									day: "numeric",
									month: "short",
									year: "numeric",
									hour: "2-digit",
									minute: "2-digit",
								});

								return (
									<Card
										key={session.id}
										className="h-full transition-all hover:shadow-md border-2 border-gray-100 hover:border-blue-200"
									>
										<CardContent className="p-5">
											<div className="flex items-start justify-between mb-3">
												<div className="flex items-center text-sm text-gray-500">
													<Clock className="h-4 w-4 mr-1.5" />
													{formattedDate}
												</div>
												{topParty && (
													<div className="text-2xl font-bold text-blue-600">
														{formatPercentage(topParty.agreement)}
													</div>
												)}
											</div>
											{topParty && (
												<div className="mb-2">
													<div className="text-sm text-gray-600 mb-1">
														Beste match:
													</div>
													<div className="font-semibold text-gray-900">
														{topParty.party.shortName}
													</div>
												</div>
											)}
											{session.data && (
												<div className="text-xs text-gray-500 mt-3 pt-3 border-t border-gray-100">
													{session.data.totalAnswers} stellingen beantwoord
												</div>
											)}
											<div className="mt-4 flex gap-2">
												<Link to={`/results/${session.id}`} className="flex-1">
													<Button
														variant="secondary"
														size="sm"
														className="w-full text-xs"
													>
														Resultaten
													</Button>
												</Link>
												<Link
													to={`/compass?session=${session.id}`}
													className="flex-1"
												>
													<Button
														variant="primary"
														size="sm"
														className="w-full text-xs"
													>
														Ga verder
													</Button>
												</Link>
											</div>
										</CardContent>
									</Card>
								);
							})}
						</div>
					</div>
				</div>
			)}

			{/* Main Tools Section */}
			<div className="container mx-auto px-4 py-16">
				<div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
					<Link to="/compass/settings" className="group block h-full">
						<Card className="h-full transition-all group-hover:shadow-md border-2 border-blue-100 hover:border-blue-200">
							<CardContent className="p-6 h-full flex flex-col">
								<div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4 group-hover:bg-blue-200 transition-colors">
									<Vote className="h-6 w-6 text-blue-600" />
								</div>
								<h3 className="font-semibold text-lg mb-2">Kieswijzer</h3>
								<p className="text-gray-600 text-sm leading-relaxed flex-grow">
									Stem zélf op moties, voor úw onderwerpen. Zie welke partijen
									het meest met u eens zijn.
								</p>
								<div className="mt-4 text-blue-600 text-sm font-medium group-hover:underline">
									Start kieswijzer →
								</div>
							</CardContent>
						</Card>
					</Link>

					<Link to="/motions" className="group block h-full">
						<Card className="h-full transition-all group-hover:shadow-md border-2 border-green-100 hover:border-green-200">
							<CardContent className="p-6 h-full flex flex-col">
								<div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4 group-hover:bg-green-200 transition-colors">
									<Search className="h-6 w-6 text-green-600" />
								</div>
								<h3 className="font-semibold text-lg mb-2">Moties</h3>
								<p className="text-gray-600 text-sm leading-relaxed flex-grow">
									Zie stemmingen per onderwerp. Zie voor úw onderwerp hoe
									partijen hebben gestemd.
								</p>
								<div className="mt-4 text-green-600 text-sm font-medium group-hover:underline">
									Bekijk moties →
								</div>
							</CardContent>
						</Card>
					</Link>

					<Link to="/party-likeness" className="group block h-full">
						<Card className="h-full transition-all group-hover:shadow-md border-2 border-orange-100 hover:border-orange-200">
							<CardContent className="p-6 h-full flex flex-col">
								<div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center mb-4 group-hover:bg-orange-200 transition-colors">
									<Users className="h-6 w-6 text-orange-600" />
								</div>
								<h3 className="font-semibold text-lg mb-2">
									Partijgelijkenissen
								</h3>
								<p className="text-gray-600 text-sm leading-relaxed flex-grow">
									Zie welke partijen qua stemgedrag op elkaar lijken, totaal en
									per onderwerp.
								</p>
								<div className="mt-4 text-orange-600 text-sm font-medium group-hover:underline">
									Bekijk gelijkenissen →
								</div>
							</CardContent>
						</Card>
					</Link>
				</div>
			</div>

			{/* Key Stats */}
			<div className="bg-gray-50 py-12">
				<div className="container mx-auto px-4 text-center">
					<div className="grid md:grid-cols-3 gap-8">
						<div>
							<div className="text-3xl font-bold text-blue-600 mb-2">
								{isLoading || !motionStatisticsData
									? "..."
									: motionStatisticsData.count.toLocaleString("nl-NL")}
							</div>
							<div className="text-gray-600">Moties geanalyseerd</div>
						</div>
						<div>
							<div className="text-3xl font-bold text-blue-600 mb-2">
								{isLoading ||
								!motionStatisticsData?.firstMotionDate ||
								!motionStatisticsData?.lastMotionDate
									? "..."
									: `${motionStatisticsData.firstMotionDate.getFullYear()}-${motionStatisticsData.lastMotionDate.getFullYear()}`}
							</div>
							<div className="text-gray-600">Periode van data</div>
						</div>
						<div>
							<div className="text-3xl font-bold text-blue-600 mb-2">100%</div>
							<div className="text-gray-600">Transparant</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}
