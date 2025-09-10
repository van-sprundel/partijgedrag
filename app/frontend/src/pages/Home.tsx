import { ArrowRight, BarChart3, Search, Users, Vote } from "lucide-react";
import { Link } from "react-router-dom";
import { useGetMotionStatistics } from "@/hooks/useGetMotionsCount";
import logo from "../assets/tug.png";
import { Button } from "../components/ui/Button";

export function HomePage() {
	const { data: motionStatisticsData, isLoading } = useGetMotionStatistics();
	return (
		<div className="min-h-screen pb-16 bg-white">
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

			{/* Main Tools Section */}
			<div className="container mx-auto px-4 py-16">
				<div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
					<Link to="/compass/settings" className="group">
						<div className="bg-white border-2 border-blue-100 rounded-lg p-6 hover:border-blue-200 hover:shadow-md transition-all">
							<div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4 group-hover:bg-blue-200 transition-colors">
								<Vote className="h-6 w-6 text-blue-600" />
							</div>
							<h3 className="font-semibold text-lg mb-2">Kieswijzer</h3>
							<p className="text-gray-600 text-sm leading-relaxed">
								Stem zélf op moties, voor úw onderwerpen. Zie welke partijen het
								meest met u eens zijn.
							</p>
							<div className="mt-4 text-blue-600 text-sm font-medium group-hover:underline">
								Start kieswijzer →
							</div>
						</div>
					</Link>

					<Link to="/moties" className="group">
						<div className="bg-white border-2 border-green-100 rounded-lg p-6 hover:border-green-200 hover:shadow-md transition-all">
							<div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4 group-hover:bg-green-200 transition-colors">
								<Search className="h-6 w-6 text-green-600" />
							</div>
							<h3 className="font-semibold text-lg mb-2">Moties</h3>
							<p className="text-gray-600 text-sm leading-relaxed">
								Zie stemmingen per onderwerp. Zie voor úw onderwerp hoe partijen
								hebben gestemd.
							</p>
							<div className="mt-4 text-green-600 text-sm font-medium group-hover:underline">
								Bekijk moties →
							</div>
						</div>
					</Link>

					<Link to="/party-focus" className="group">
						<div className="bg-white border-2 border-purple-100 rounded-lg p-6 hover:border-purple-200 hover:shadow-md transition-all">
							<div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4 group-hover:bg-purple-200 transition-colors">
								<BarChart3 className="h-6 w-6 text-purple-600" />
							</div>
							<h3 className="font-semibold text-lg mb-2">Partijfocus</h3>
							<p className="text-gray-600 text-sm leading-relaxed">
								Bekijk de onderwerpen waar een partij de meeste moties over
								indient.
							</p>
							<div className="mt-4 text-purple-600 text-sm font-medium group-hover:underline">
								Bekijk focus →
							</div>
						</div>
					</Link>

<Link to="/party-likeness" className="group">
						<div className="bg-white border-2 border-orange-100 rounded-lg p-6 hover:border-orange-200 hover:shadow-md transition-all">
							<div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center mb-4 group-hover:bg-orange-200 transition-colors">
								<Users className="h-6 w-6 text-orange-600" />
							</div>
							<h3 className="font-semibold text-lg mb-2">
								Partijgelijkenissen
							</h3>
							<p className="text-gray-600 text-sm leading-relaxed">
								Zie welke partijen qua stemgedrag op elkaar lijken, totaal en
								per onderwerp.
							</p>
							<div className="mt-4 text-orange-600 text-sm font-medium group-hover:underline">
								Bekijk gelijkenissen →
							</div>
						</div>
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
								2008-2025
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
