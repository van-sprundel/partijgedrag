import { useLocation } from 'react-router';
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink, BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from '~/components/ui/breadcrumb';
import { capitalize } from '~/lib/utils';
import React from 'react';

export function DashboardBreadcrumbs() {
    const location = useLocation();
    const breadcrumbs = location.pathname.split('/').filter(Boolean);

    function generateBreadcrumbs() {
        return breadcrumbs.map((breadcrumb, index) => {
            if (index === breadcrumbs.length - 1) {
                return <BreadcrumbPage key={breadcrumb}>{capitalize(breadcrumb)}</BreadcrumbPage>
            }
            return <>
                <BreadcrumbItem key={breadcrumb} className={'hidden md:block'}>
                    <BreadcrumbLink  href={`/${breadcrumbs.slice(0, index + 1).join('/')}`}>{capitalize(breadcrumb)}</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator key={`${breadcrumb}-separator`} className={'hidden md:block '} />
            </>

        })
    }

    return (
        <Breadcrumb>
            <BreadcrumbList>
                {generateBreadcrumbs()}
            </BreadcrumbList>
        </Breadcrumb>
    )
}