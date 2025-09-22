import { Github } from "lucide-react";
import cysoLogo from "../../assets/cyso.svg";

export function Footer() {
	return (
		<footer className="bg-white border-t shadow-sm mt-auto">
			<div className="container mx-auto px-4 py-6">
				<div className="flex flex-col md:flex-row justify-between items-center gap-4">
					{/* Left side - Copyright and CYSO */}
					<div className="flex items-center gap-4 text-sm text-gray-600">
						<span>© 2012-2025 Elwin Oost,</span>
						<a
							href="https://solveware.nl"
							target="_blank"
							rel="noopener noreferrer"
							className="hover:text-primary-600 transition-colors"
						>
							Solveware
						</a>
						<span>. Partijgedrag wordt gehost door</span>
						<a
							href="https://cyso.com"
							target="_blank"
							rel="noopener noreferrer"
							className="flex items-center gap-2 hover:opacity-80 transition-opacity"
						>
							<img src={cysoLogo} alt="CYSO" className="h-6" />
						</a>
					</div>

					{/* Right side - GitHub link */}
					<div className="flex items-center gap-2">
						<span className="text-sm text-gray-600">Open source op</span>
						<a
							href="https://github.com/van-sprundel/partijgedrag"
							target="_blank"
							rel="noopener noreferrer"
							className="flex items-center gap-2 text-gray-600 hover:text-primary-600 transition-colors"
						>
							<Github className="h-5 w-5" />
							<span className="text-sm font-medium">GitHub</span>
						</a>
					</div>
				</div>
			</div>
		</footer>
	);
}
