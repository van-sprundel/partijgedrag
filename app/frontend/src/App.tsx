import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { CompassPage } from "./pages/Compass";
import { CompassSettingsPage } from "./pages/CompassSettings";
import { HomePage } from "./pages/Home";
import { PartyLikenessPage } from "./pages/PartyLikeness";
import { ResultsPage } from "./pages/Results";

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			staleTime: 5 * 60 * 1000, // 5 minutes
			retry: 1,
		},
	},
});

function App() {
	return (
		<QueryClientProvider client={queryClient}>
			<BrowserRouter>
				<div className="min-h-screen bg-gray-50">
					<Routes>
						<Route path="/" element={<HomePage />} />
						<Route path="/compass/settings" element={<CompassSettingsPage />} />
						<Route path="/compass" element={<CompassPage />} />
						<Route path="/party-likeness" element={<PartyLikenessPage />} />
						<Route path="/results/:sessionId" element={<ResultsPage />} />
						<Route path="*" element={<NotFoundPage />} />
					</Routes>
				</div>
			</BrowserRouter>
			{import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
		</QueryClientProvider>
	);
}

function NotFoundPage() {
	return (
		<div className="min-h-screen bg-gray-50 flex items-center justify-center">
			<div className="text-center">
				<h1 className="text-6xl font-bold text-gray-300 mb-4">404</h1>
				<h2 className="text-2xl font-semibold text-gray-900 mb-4">
					Pagina niet gevonden
				</h2>
				<p className="text-gray-600 mb-8">
					De pagina die je zoekt bestaat niet.
				</p>
				<a
					href="/"
					className="inline-flex items-center px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 transition-colors"
				>
					Terug naar home
				</a>
			</div>
		</div>
	);
}

export default App;
