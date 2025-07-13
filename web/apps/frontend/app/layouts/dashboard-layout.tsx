import { Link, Outlet, redirect, useLoaderData, useLocation, useNavigate } from 'react-router';
import { authLoader } from '~/loaders/auth-loader';
import type { Route } from '../../.react-router/types/app/routes/+types/_index';
import { Button } from '~/components/ui/button';
import { HouseIcon, LayoutDashboardIcon, UserIcon } from 'lucide-react';
import {  PanelLeft } from 'lucide-react';
import { cx } from 'class-variance-authority';
import { Separator } from '~/components/ui/separator';
import { DashboardBreadcrumbs } from '~/components/dashboard/DashboardBreadcrumbs';
import { DashboardRoute } from '~/components/dashboard/DashboardRoute';

export async function loader(loaderArgs: Route.LoaderArgs) {
    const user = await authLoader(loaderArgs);
    const searchParams = new URL(loaderArgs.request.url).searchParams;
    if (!user) {
        return redirect('/auth/login');
    }
    const expanded = searchParams.get('expanded') === 'true';
    return {
        user,
        expanded,
    };
}


// generate this based on the routes
const routes: DashboardRoute[] = [
    {
        path: 'dashboard',
        name: 'Dashboard',
        icon: LayoutDashboardIcon,
        isRoot: true,
    },
    {
        path: 'dashboard/profile',
        name: 'Profile',
        icon: UserIcon,
    },
    {
        path: 'dashboard/sub-route',
        name: 'Sub Route',
        icon: HouseIcon,
        children: [
            {
                path: 'dashboard/sub-route/sub-sub-route',
                name: 'Sub Sub Route',
                // icon: HouseIcon,
            },
            {
                path: 'dashboard/sub-route/test',
                name: 'Test',
                // icon: HouseIcon,
            },
        ],
    },
];


export default function DashboardLayout() {
    const loaderData = useLoaderData()
    const expanded = loaderData.expanded;
    const location = useLocation();
    const navigate = useNavigate();

    const toggleSidebar = () => {
        const searchParams = new URLSearchParams();
        searchParams.set('expanded', (!expanded).toString());
        navigate(`${location.pathname}?${searchParams.toString()}`, { replace: true });
    };

    return <div className={'w-max-screen h-screen  bg-secondary overflow-hidden grid grid-cols-[auto_1fr] md:px-2 transition-all ease-in-out duration-700'}>
        <aside
            className={cx('transition-all h-s  ease-in-out bg-secondary  grid grid-rows-[auto_1fr_auto]  h-screen md:h-full items-start sticky top-0 self-auto overflow-hidden text-nowrap py-4 w-[0] md:w-12 px-2 justify-center  ', expanded ? 'z-20 w-[80vw] md:flex-col md:w-60 px-4 justify-stretch' : ' ')}>
            <div className={cx("h-12 justify-center items-center flex", expanded && 'justify-start')}>
                <Link to={'/'} className={'w-fit '}>
                    <span className={'text-xl font-bold text-primary-gradient'}>FTS</span>
                </Link>
            </div>
            <ul className={'flex flex-col grow '}>
                {routes.map(route => (<DashboardRoute key={route.path} route={route} expanded={expanded} />))}
            </ul>
            <div className={'self-center'}>
                othershit
            </div>
        </aside>
        <main
            className={'overflow-hidden bg-background md:rounded-2xl md:ml-4 h-screen md:h-auto  w-screen md:w-auto  fixed md:relative flex l flex-1 flex-col  md:peer-data-[variant=inset]:m-2 md:peer-data-[state=collapsed]:peer-data-[variant=inset]:ml-2 md:peer-data-[variant=inset]:ml-0 md:peer-data-[variant=inset]:rounded-xl md:peer-data-[variant=inset]:shadow  md:my-4'}
           >
            {expanded && <div className={cx('md:hidden w-screen md:w-full h-screen md:h-full fixed top-0 left-0')}  onClick={() => {
                toggleSidebar();
            }}></div>}
            <div className={'rounded-2xl h-full '}>
                <div className={'bg-background px-6 gap-1 items-center w-full flex border-b h-12 sticky top-0 '}>
                    <Button variant="ghost" className={'p-1'} size="icon" onClick={toggleSidebar} asChild>
                        <PanelLeft
                            className={cx('inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 hover:bg-accent hover:text-accent-foreground h-7 w-7 -ml-1')} />
                    </Button>
                    <Separator orientation="vertical"
                               className={'shrink-0 bg-border h-full w-[1px] mx-2 data-[orientation=vertical]:h-4'} />
                    <DashboardBreadcrumbs/>
                </div>
                <div className={'rounded-2xl px-6 py-8  h-full overflow-x-hidden overflow-y-auto'}>
                    <Outlet />

                </div>
                <div className={'h-4 sticky bottom-0 w-full bg-background'} >
                </div>
            </div>
        </main>
    </div>;
}




