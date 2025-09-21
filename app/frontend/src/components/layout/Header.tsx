import { Link } from "react-router-dom";
import logo from "../../assets/tug.png";
import { Button } from "../ui/Button";

export function Header() {
	return (
		<header className="bg-white shadow-sm border-b">
			<div className="container mx-auto px-4 py-3 flex justify-between items-center">
				<Link
					to="/"
					className="text-2xl font-bold text-gray-900 flex justify-center gap-2"
				>
					<img src={logo} alt="Partijgedrag Logo" className="size-8" />
					Partijgedrag
				</Link>
				<nav className="hidden md:flex items-center space-x-2">
					<Link to="/compass/settings">
						<Button variant="ghost">Stemwijzer</Button>
					</Link>
					<Link to="/party-likeness">
						<Button variant="ghost">Partijgelijkenis</Button>
					</Link>
					{/*<Link to="/motions">
						<Button variant="ghost">Moties</Button>
					</Link>*/}
					<Link to="/about">
						<Button variant="ghost">Over ons</Button>
					</Link>
				</nav>
			</div>
		</header>
	);
}
