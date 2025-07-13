import { Link, useLocation } from 'react-router';
import React, { useEffect } from 'react';
import { cx } from 'class-variance-authority';
import { Button } from '~/components/ui/button';
import { ChevronUp, type LucideIcon } from 'lucide-react';

export type DashboardRoute = {
    path: string;
    name: string;
    icon?: LucideIcon;
    children?: DashboardRoute[];
    isRoot?: boolean;
}


export function DashboardRoute({ route, expanded }: { route: DashboardRoute, expanded?: boolean }) {
    const location = useLocation();
    const search = location.search; // includes '?open=true'

    const isSelected = `/${route.path}` === location.pathname;
    const isSelectedParent = location.pathname.includes(`/${route.path}`) && (!route.isRoot || route.isRoot && `/${route.path}` === location.pathname);
    const [open, setOpen] = React.useState(isSelectedParent);
    const hasChildren = route.children && route.children.length > 0;


    useEffect(() => {
        setOpen(isSelectedParent);
    }, [isSelectedParent]);

    const toggleDropdown = (e: React.MouseEvent<HTMLElement>) => {
        if(isSelectedParent) return;
        e.preventDefault();
        e.stopPropagation();
        setOpen((prev) => {
            return !prev;
        });
    };
    return (
        <>
            <li className={cx('  flex-row flex-1 flex w-full items-end text-nowrap overflow-hidden hover:bg-background rounded-sm cursor-pointer', isSelectedParent && 'text-primary-gradient', expanded ? '' : 'my-2') } onClick={(e) =>toggleDropdown(e)}>
                <Link to={`${route.path}${search}`} className={cx('grid  items-center h-10  text-nowrap overflow-hidden w-full ease-in-out transition-all duration-[5000]', expanded ? 'gap-2 grid-cols-[auto_1fr_auto] px-2 justify-start' : 'grid-cols-[1fr_0fr_0fr] place-items-center justify-center h-8')}>
                    {route.icon && <route.icon className={cx(' w-5 h-5', !expanded && 'w-4 h-4')} />}
                    <div className={cx(!expanded && 'overflow-hidden text-nowrap w-0', isSelected && !hasChildren && 'underline')}>
                        {route.name}
                    </div>
                    {
                        expanded && hasChildren && <div className={' flex items-center self-end'}>
                            <Button className={' focus-visible:ring-0 focus-visible:ring-offset-0 hover:bg-transparent'} variant="ghost" size="icon" onClick={toggleDropdown}>
                                <ChevronUp
                                    className={cx(hasChildren && !open && 'rotate-[180deg]', 'ease-in-out transition', 'size-3')} />
                            </Button>
                        </div>
                    }
                </Link>
            </li>
            <div className={cx(' pl-4 grid ease-in-out transition-all duration-700', expanded && open ? 'grid-rows-[1fr] ' : 'grid-rows-[0fr] ')}>
                <div className={'overflow-hidden'}>
                    {hasChildren && [{path: route.path, name: route.name, isRoot: true}, ...route.children ?? []].map(child => <DashboardRoute key={child.path} route={child} expanded={open} />)}
                </div>
            </div>
        </>
    );
}
