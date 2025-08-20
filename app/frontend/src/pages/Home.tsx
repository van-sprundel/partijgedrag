import { ArrowRight, Vote } from "lucide-react";
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
					<Link to="/compass">
						<Button size="lg" className="text-lg px-8 py-4">
							Start de stemwijzer
							<ArrowRight className="ml-2 h-5 w-5" />
						</Button>
					</Link>
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
					<Link to="/compass">
						<Button size="lg" className="text-lg px-12 py-4">
							Begin nu
							<ArrowRight className="ml-2 h-5 w-5" />
						</Button>
					</Link>
				</div>
			</div>
		</div>
	);
}
