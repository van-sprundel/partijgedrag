import type { Route } from '../+types/_index';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { authClient } from '~/clients';
import { Button } from '~/components/ui/button';
import PublicFormLayout from '~/layouts/public-form-layout';
import CForm, { type CField } from '~/components/auth/c-form';
import React from 'react';

export async function action(actionArgs:  Route.ActionArgs) {
  let formData = await actionArgs.request.formData();
  const email = formData.get('email') as string;
  const url = new URL(actionArgs.request.url);
  if (!email) {
    return {
      status: 400,
      body: {
        message: 'Email is required',
      },
    };
  }

  const result = await authClient.forgotPassword({
    params: { email },
    query: { redirectUrl: `${url.origin}/auth/reset-password` },
  });

  return result
}

const initialForgotPasswordParam = {
  email: '',
}

// TODO: wrap form in own component with zod validations :)
export default function Page() {

  const forgotPasswordHeader = (
    <>
      <CardTitle>Send Password reset email</CardTitle>
      <CardDescription>
        Enter your email below to reset your password
      </CardDescription>
    </>
  )

  const forgotPasswordFields: CField[] = [
    {
      name: 'email',
      type: 'email',
      label: 'Email',
      placeholder: 'm@example.com',
      required: true,
    },
  ];

  const forgotPasswordButtons = (
    <Button type="submit" className="w-full rounded-xl">
      Send
    </Button>
  )

  return (
    <PublicFormLayout cardHeader={forgotPasswordHeader} cardContent={
      <>
        <CForm initialBody={initialForgotPasswordParam} fields={forgotPasswordFields} buttons={forgotPasswordButtons} />
        <div className="mt-4 text-center text-sm">
          You're email will be only valid for 15 minutes
          <br />
          Don't have an account?{' '}
          <a href="/auth/register" className="underline underline-offset-4">
            Sign up
          </a>
        </div>
      </>
    } />
  );
}
