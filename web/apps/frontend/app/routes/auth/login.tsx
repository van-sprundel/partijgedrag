import React, { useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { Form, Link, redirect, useActionData } from 'react-router';
import { Button } from '~/components/ui/button';
import { authClient, type AuthClientResponseShapes } from '~/clients';
import { nonAuthLoader } from '~/loaders/auth-loader';
import type { Route } from '../../../.react-router/types/app/+types/root';
import PublicFormLayout from '~/layouts/public-form-layout';
import CForm, { type CField } from '~/components/auth/c-form';


const initialLoginBody = {
  email: '',
  password: '',
};

export async function loader(loaderArgs: Route.LoaderArgs) {
  return await nonAuthLoader(loaderArgs);
}

export async function action({ request, context }:  Route.ActionArgs) {
  let formData = await request.formData();
  const email = formData.get('email') as string;
  const password = formData.get('password') as string;
  if (!email || !password) {
    return {
      status: 400,
      body: {
        message: 'Email and password are required',
      }
    };
  }
  const result = await authClient.login({
    body: {
      email,
      password,
    },
  })
  if (!result.status || result.status !== 200) {
    return result;
  }

  return redirect(`/dashboard`, {
    status: 200,
    headers: {
      'Set-Cookie': result.headers.get('Set-Cookie') ?? '',
    },
  });
}


export default function Page({}: Route.ComponentProps) {
  const loginHeader = (
    <>
      <CardTitle>Login to your account</CardTitle>
      <CardDescription>
        Enter your email below to login to your account
      </CardDescription>
    </>
  )

  const loginFields: CField[] = [
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
      rightContent: <Link to="/auth/forgot-password">Forgot password?</Link>,
    },
  ];

  const loginButtons = (
    <Button type="submit" className="w-full rounded-xl">
      Login
    </Button>
  )

  return (
    <PublicFormLayout cardHeader={loginHeader} cardContent={
     <>
     <CForm initialBody={initialLoginBody} fields={loginFields} buttons={loginButtons} />
     <div className="mt-4 text-center text-sm">
       Don't have an account?{' '}
       <a href="/auth/register" className="underline underline-offset-4">
         Sign up
       </a>
    </div>
     </>
    } />

  );
}
