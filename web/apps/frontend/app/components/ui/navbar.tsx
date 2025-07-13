import { Button } from '~/components/ui/button';
import { Menu } from "lucide-react";
import { useEffect, useState } from 'react';
import { Link, useFetcher, useLoaderData } from 'react-router';

export default function Navbar() {
  const [menuOpen, setMenuOpen] = useState(false);
  const user = useLoaderData();

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 backdrop-blur-md  shadow-lg border-b border-grad">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center">
              <a href="/">
                <h2 className="text-2xl font-bold bg-gradient-to-br from-primary to-primary/40 bg-clip-text text-transparent">
                  FTS
                </h2>
                </a>
          </div>
          {!user &&
            (            <div className="flex items-center space-x-4">
            <Button variant="ghost" className="text-sm" asChild>
              <Link to="/auth/login">Login</Link>
            </Button>
            <Button className="text-sm bg-gradient-to-br from-primary to-primary/40 hover:opacity-90" asChild>
              <Link to="/auth/register">Register</Link>
            </Button>
          </div>)
          }
          {user &&
            (<div className="flex items-center space-x-4">
              <Button variant="destructive" className="text-sm bg-gradient-to-r from-red-600 to-red-600/40 hover:opacity-90" asChild>
                <Link to="/auth/logout">Logout</Link>
              </Button>
              <Button className="text-sm bg-gradient-to-r from-primary to-primary/40 hover:opacity-90" asChild>
                <Link to="/dashboard">Open App</Link>
              </Button>
            </div>)
          }
        </div>
      </div>
    </nav>
  );
}
