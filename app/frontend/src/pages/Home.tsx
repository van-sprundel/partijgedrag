import { ArrowRight, BarChart3, Vote } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "../components/ui/Button";
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from "../components/ui/Card";

export function HomePage() {
	return (
		<div className="min-h-screen bg-gradient-to-b from-primary-50 to-white">
			<div className="container mx-auto px-4 py-16">
				{/* Hero Section */}
				<div className="text-center mb-16">
					<div className="flex justify-center mb-8">
						<div className="p-4 bg-primary-100 rounded-full">
							<Vote className="h-16 w-16 text-primary-600" />
						</div>
					</div>
					<h1 className="text-5xl font-bold text-gray-900 mb-6">
						Ontdek jouw politieke match
					</h1>
					<p className="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
						Beantwoord stellingen over belangrijke onderwerpen en ontdek welke
						Nederlandse politieke partij het beste aansluit bij jouw
						standpunten.
					</p>
					<div className="flex flex-col sm:flex-row gap-4 justify-center">
						<Link to="/compass">
							<Button size="lg" className="text-lg px-8 py-4">
								Start de stemwijzer
								<ArrowRight className="ml-2 h-5 w-5" />
							</Button>
						</Link>
						<Link to="/party-likeness">
							<Button
								variant="secondary"
								size="lg"
								className="text-lg px-8 py-4"
							>
								Partijgelijkenissen
								<BarChart3 className="ml-2 h-5 w-5" />
							</Button>
						</Link>
					</div>
				</div>

				{/* Features */}
				<div className="grid md:grid-cols-3 gap-8 mb-16">
					<Card className="text-center hover:shadow-lg transition-shadow">
						<CardHeader>
							<div className="mx-auto w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
								<span className="text-2xl">📊</span>
							</div>
							<CardTitle className="text-xl">Gebaseerd op echte data</CardTitle>
						</CardHeader>
						<CardContent>
							<p className="text-gray-600">
								Onze stellingen zijn gebaseerd op daadwerkelijke moties en
								stemmingen uit de Tweede Kamer.
							</p>
						</CardContent>
					</Card>

					<Card className="text-center hover:shadow-lg transition-shadow">
						<CardHeader>
							<div className="mx-auto w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
								<span className="text-2xl">⚡</span>
							</div>
							<CardTitle className="text-xl">Snel en eenvoudig</CardTitle>
						</CardHeader>
						<CardContent>
							<p className="text-gray-600">
								In slechts 20 vragen krijg je inzicht in welke partijen het
								beste aansluiten bij jouw standpunten.
							</p>
						</CardContent>
					</Card>

					<Card className="text-center hover:shadow-lg transition-shadow">
						<CardHeader>
							<div className="mx-auto w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
								<span className="text-2xl">🔍</span>
							</div>
							<CardTitle className="text-xl">
								Gedetailleerde resultaten
							</CardTitle>
						</CardHeader>
						<CardContent>
							<p className="text-gray-600">
								Zie niet alleen welke partijen bij je passen, maar ook waarom en
								hoe ze hebben gestemd.
							</p>
						</CardContent>
					</Card>
				</div>

				{/* Party Likeness Section */}
				<div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-2xl p-8 mb-16 border border-blue-100">
					<div className="text-center mb-8">
						<div className="flex justify-center mb-6">
							<div className="p-4 bg-blue-100 rounded-full">
								<BarChart3 className="h-12 w-12 text-blue-600" />
							</div>
						</div>
						<h2 className="text-3xl font-bold text-gray-900 mb-4">
							Partijgelijkenissen
						</h2>
						<p className="text-lg text-gray-700 max-w-3xl mx-auto mb-6">
							Ontdek welke politieke partijen het meest overeenkomen in hun
							stemgedrag. Gebaseerd op echte stemmingen in de Tweede Kamer.
						</p>
						<Link to="/party-likeness">
							<Button
								size="lg"
								variant="secondary"
								className="bg-white hover:bg-blue-50 text-blue-600 border-blue-200"
							>
								Bekijk partijgelijkenissen
								<ArrowRight className="ml-2 h-5 w-5" />
							</Button>
						</Link>
					</div>

					<div className="grid md:grid-cols-2 gap-6 text-center">
						<div className="bg-white/50 rounded-lg p-4">
							<div className="text-2xl font-bold text-blue-600 mb-2">📊</div>
							<h3 className="font-semibold mb-2">Vergelijk stemgedrag</h3>
							<p className="text-sm text-gray-600">
								Zie welke partijen vaak hetzelfde stemmen op belangrijke moties
							</p>
						</div>
						<div className="bg-white/50 rounded-lg p-4">
							<div className="text-2xl font-bold text-blue-600 mb-2">🎯</div>
							<h3 className="font-semibold mb-2">Objectieve analyse</h3>
							<p className="text-sm text-gray-600">
								Gebaseerd op daadwerkelijke stemmingen, niet op
								verkiezingsprogramma's
							</p>
						</div>
					</div>
				</div>

				{/* How it works */}
				<div className="bg-white rounded-2xl p-8 shadow-sm">
					<h2 className="text-3xl font-bold text-center mb-8 text-gray-900">
						Hoe werkt het?
					</h2>
					<div className="grid md:grid-cols-4 gap-6">
						<div className="text-center">
							<div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
								<span className="text-2xl font-bold text-primary-600">1</span>
							</div>
							<h3 className="font-semibold mb-2">Beantwoord stellingen</h3>
							<p className="text-gray-600 text-sm">
								Geef aan of je het eens of oneens bent met verschillende
								politieke stellingen.
							</p>
						</div>
						<div className="text-center">
							<div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
								<span className="text-2xl font-bold text-primary-600">2</span>
							</div>
							<h3 className="font-semibold mb-2">Vergelijk met partijen</h3>
							<p className="text-gray-600 text-sm">
								We vergelijken jouw antwoorden met hoe partijen werkelijk hebben
								gestemd.
							</p>
						</div>
						<div className="text-center">
							<div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
								<span className="text-2xl font-bold text-primary-600">3</span>
							</div>
							<h3 className="font-semibold mb-2">Zie je match</h3>
							<p className="text-gray-600 text-sm">
								Krijg een overzicht van welke partijen het beste bij je passen.
							</p>
						</div>
						<div className="text-center">
							<div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
								<span className="text-2xl font-bold text-primary-600">4</span>
							</div>
							<h3 className="font-semibold mb-2">Verdiep je</h3>
							<p className="text-gray-600 text-sm">
								Bekijk gedetailleerde informatie over waarom partijen zo hebben
								gestemd.
							</p>
						</div>
					</div>
				</div>

				{/* CTA */}
				<div className="text-center mt-16">
					<h2 className="text-3xl font-bold mb-4 text-gray-900">
						Klaar om te beginnen?
					</h2>
					<p className="text-xl text-gray-600 mb-8">
						Het duurt maar 5 minuten en het is volledig gratis.
					</p>
					<div className="flex flex-col sm:flex-row gap-4 justify-center">
						<Link to="/compass">
							<Button size="lg" className="text-lg px-12 py-4">
								Begin nu
								<ArrowRight className="ml-2 h-5 w-5" />
							</Button>
						</Link>
						<Link to="/party-likeness">
							<Button
								variant="secondary"
								size="lg"
								className="text-lg px-8 py-4"
							>
								Of bekijk partijgelijkenissen
								<BarChart3 className="ml-2 h-5 w-5" />
							</Button>
						</Link>
					</div>
				</div>
			</div>
		</div>
	);
}
