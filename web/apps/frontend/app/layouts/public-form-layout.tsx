import { cn } from '~/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import React from 'react';

export type PublicFormLayoutProps = {
  cardContent?: React.ReactNode;
  cardHeader?: React.ReactNode;
};

export default function PublicFormLayout({ cardContent, cardHeader }: PublicFormLayoutProps) {
  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm">
        <div className={cn('flex flex-col gap-6')} >
          <Card>
            <CardHeader>
              {cardHeader}
            </CardHeader>
            <CardContent>
              {cardContent}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}