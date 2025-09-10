import { useGetMotionStatistics } from "@/hooks/useGetMotionsCount";

export function AboutPage() {
	const { data: motionStatisticsData, isLoading } = useGetMotionStatistics();
	return (
		<div className="max-w-2xl mx-auto px-4 py-12 sm:px-6 lg:max-w-5xl xl:max-w-6xl lg:px-8">
			{/* Header Section */}
			<div className="mb-12">
				<h1 className="text-4xl font-bold text-gray-900 mb-8">Over ons</h1>

				<div className="prose prose-lg max-w-none text-gray-700 space-y-6">
					<p>
						Partijgedrag is gemaakt door <strong>Elwin Oost</strong>, eigenaar
						van Solveware en security architect van SDB Groep.
					</p>
					<p>
						Deze site is ontstaan uit mijn wens meer duidelijkheid te krijgen,
						en aan geïnteresseerden te geven, over het werkelijke gedrag van
						partijen. Ik heb daartoe geautomatiseerd het stemgedrag van de
						partijen ontleed en makkelijk doorzoekbaar gemaakt.
					</p>
					<p>
						De site is geheel onafhankelijk. Er is ook geen gebruik gemaakt van
						subsidie. Dank aan alle vrienden die hebben meegedacht, en Cyso voor
						het hosten van de site.
					</p>
				</div>
			</div>

			{/* Actualiteit Section */}
			<div className="mb-12">
				<h2 className="text-2xl font-bold mb-4">Actualiteit</h2>
				<p className="text-lg text-gray-700">
					De site omvat nu{" "}
					{!isLoading && (
						<>
							<strong>{motionStatisticsData?.count}</strong>, van{" "}
{motionStatisticsData?.firstMotionDate?.toLocaleDateString("nl-NL")}{" "}
tot{" "}
{motionStatisticsData?.lastMotionDate?.toLocaleDateString("nl-NL")}
							.
						</>
					)}
				</p>
			</div>

			{/* FAQ Section */}
			<div className="mb-12">
				<h2 className="text-2xl font-bold text-gray-900 mb-8">
					Veelgestelde vragen
				</h2>

				<div className="space-y-8">
					<div>
						<h3 className="text-xl font-semibold text-gray-900 mb-3">
							Waar komt de informatie vandaan?
						</h3>
						<p className="text-gray-700 leading-relaxed">
							Deze is met tekstanalyse afgeleid van de handelingen uit de
							openbare informatie van de officiële bekendmakingen overheidssite.
						</p>
					</div>

					<div>
						<h3 className="text-xl font-semibold text-gray-900 mb-3">
							Waarom geen amendementen en wetsvoorstellen?
						</h3>
						<p className="text-gray-700 leading-relaxed">
							Deze zijn veelal niet leesbaar zonder de gerelateerde wetteksten,
							waardoor de strekking onduidelijk blijft. Moties zijn veel
							directer.
						</p>
					</div>

					<div>
						<h3 className="text-xl font-semibold text-gray-900 mb-3">
							Waar zijn de eenmansfracties?
						</h3>
						<p className="text-gray-700 leading-relaxed">
							Partijgedrag bevat alleen informatie over partijen die minimaal
							een kabinetsperiode hebben volgemaakt. De eenmansfracties zijn pas
							recent afgesplitst; er is nog geen goed beeld te vormen van hun
							stemgedrag.
						</p>
					</div>

					<div>
						<h3 className="text-xl font-semibold text-gray-900 mb-3">
							Dienen alle partijen even veel moties in?
						</h3>
						<p className="text-gray-700 leading-relaxed">
							Hoewel het parlement dualisme nastreeft, dienen de
							regeringspartijen gemiddeld minder moties in dan de oppositie. U
							kunt de verhoudingen vinden op de partijfocus pagina.
						</p>
					</div>

					<div>
						<h3 className="text-xl font-semibold text-gray-900 mb-3">
							Hoe worden hoofdelijke moties geteld?
						</h3>
						<p className="text-gray-700 leading-relaxed">
							Hiervoor wordt de stem van de lijsttrekker genomen. Indien deze
							niet aanwezig is, wordt de meerderheid van stemmen door
							partijgenoten genomen.
						</p>
					</div>

					<div>
						<h3 className="text-xl font-semibold text-gray-900 mb-3">
							Waarom stemt een partij tegen zijn idealen?
						</h3>
						<p className="text-gray-700 leading-relaxed">
							Soms verduidelijkt een partij dit met een stemverklaring. Deze zal
							dan in de bijbehorende handeling staan; druk daarvoor op de knop{" "}
							<em>Toon handeling</em> onder de motie. Zo kan een partij tegen
							een motie "Verlaag X 5%" stemmen, omdat er een motie "Verlaag X
							10%" volgt waar de partij voor wil stemmen. Anders heeft u
							mogelijk een punt gevonden waarvoor diens beleid niet strookt met
							diens genoemde idealen.
						</p>
					</div>

					<div>
						<h3 className="text-xl font-semibold text-gray-900 mb-3">
							Kan mijn (politieke) banner op de site?
						</h3>
						<p className="text-gray-700 leading-relaxed">
							Nee, sorry, we willen de site geheel onafhankelijk houden.
						</p>
					</div>
				</div>
			</div>

			{/* Privacy Section */}
			<div className="mb-12">
				<h2 className="text-2xl font-bold text-gray-900 mb-6">Privacy</h2>

				<div className="prose max-w-none text-gray-700 space-y-4">
					<p>
						Uw persoonlijke gebruik van deze site blijft privé, simpel. Wij
						vragen nergens gegevens van u.
					</p>

					<div>
						<p className="font-semibold mb-3">
							De enige zaken die geregistreerd worden:
						</p>
						<ul className="space-y-2 ml-6">
							<li className="list-disc">
								<strong>Standaardlogging</strong> welke IP adressen welke
								pagina's opvragen. Deze wordt nooit met derden gedeeld tenzij
								dit wettelijk verplicht zou zijn of misbruik van de site zou
								worden geconstateerd. Hier is geen partijvoorkeur uit af te
								leiden. Na 15 dagen wordt deze gemist.
							</li>
							<li className="list-disc">
								<strong>Verder: Niets.</strong>
							</li>
						</ul>
					</div>

					<p>
						Sites van derden (zoals reclame-, social networking en
						trackingsites) kunnen u niet via deze site volgen ten zij u zelf
						kiest naar zo een site te gaan; deze zijn afgeschermd.
					</p>

					<p>
						Indien u cookies accepteert, worden deze gebruikt om uw voorkeuren
						en eventueel in de kieswijzer uitgebrachte stemmen te onthouden. We
						gebruiken een cookie met willekeurig nummer om te zien hoe de site
						gebruikt wordt. Alle cookies worden uitsluitend bij directe
						interactie met de site gebruikt.
					</p>
				</div>
			</div>

			{/* Disclaimer Section */}
			<div className="border-t border-gray-200 pt-8">
				<h2 className="text-2xl font-bold text-gray-900 mb-4">Disclaimer</h2>

				<div className="text-gray-700 space-y-3">
					<p>
						We hebben ons best gedaan de stemgegevens zo zorgvuldig mogelijk te
						vullen, maar het kan zijn dat er fouten in staan. <br />
						Tip ons op{" "}
						<code className="bg-gray-100 px-2 py-1 rounded text-sm">
							ln.erawev @los@tsoo@.m.e
						</code>
						.
					</p>
					<p>
						Aan de inhoud van deze site kunnen geen rechten worden ontleend.
					</p>
				</div>
			</div>
		</div>
	);
}
