import { authClient } from '~/clients';
import { Form, redirect } from 'react-router';
import type { Route } from '../+types/_index';
import { cn } from '~/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { Button } from '~/components/ui/button';
import CForm, { type CField } from '~/components/auth/c-form';
import PublicFormLayout from '~/layouts/public-form-layout';
import React from 'react';
import { nonAuthLoader } from '~/loaders/auth-loader';


export async function loader(loaderArgs: Route.LoaderArgs) {
  return await nonAuthLoader(loaderArgs);
}

export async function action({ request, context }: any) {
  let formData = await request.formData();

  const email = formData.get('email') as string;
  const password = formData.get('password') as string;

  if (!email || !password) {
    return {
      status: 500,
      body: {
        message: 'Email and password are required',
      }
    };
  }
  const result = await authClient.register({
    body: {
      email,
      password,
    },
  });
  if (!result.status || result.status !== 200) {
    return result;
  }
  return redirect(`/dashboard`, {
    headers: {
      'Set-Cookie': result.headers.get('Set-Cookie') ?? '',
    },
  });
}

const initialRegisterBody = {
  email: '',
  password: '',
  confirmPassword: '',
};

export default function Page() {
  const registerHeader = (
    <>
      <CardTitle>Register an account</CardTitle>
      <CardDescription>
        Enter your email below to login to your account
      </CardDescription>
    </>
  )

  const registerFields: CField[] = [
    {
      name: 'email',
      type: 'email',
      label: 'Email',
      placeholder: 'm@example.com',
      required: true,
    },
    {
      name: 'password',
      type: 'password',
      label: 'Password',
      placeholder: 'Password',
      required: true,
    },
    {
      name: 'confirmPassword',
      type: 'password',
      label: 'Confirm Password',
      placeholder: 'Confirm Password',
      required: true,
    }
  ];

  const registerButtons = (
    <Button type="submit" className="w-full rounded-xl">
      Register
    </Button>
  )

  return (
    <PublicFormLayout cardHeader={registerHeader} cardContent={
      <>
        <CForm initialBody={initialRegisterBody} fields={registerFields} buttons={registerButtons} />

        <div className="mt-4 text-center text-sm">
          Already have an account?{' '}
          <a href="/auth/login" className="underline underline-offset-4">
            Sign in
          </a>
        </div>
      </>} />

  );
}
