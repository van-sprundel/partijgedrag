import CForm, { type CField } from '~/components/auth/c-form';
import PublicFormLayout from '~/layouts/public-form-layout';
import React from 'react';
import { Button } from '~/components/ui/button';
import { CardDescription, CardTitle } from '~/components/ui/card';
import { Link, redirect, useLoaderData, useSearchParams } from 'react-router';
import type { Route } from '../../../.react-router/types/app/+types/root';
import { authClient, type AuthClientResponseShapes } from '~/clients';

export async function loader({ request}: Route.LoaderArgs) {
  const url = new URL(request.url);
  const searchParams = url.searchParams;
  const token = searchParams.get('token');
  if (!token) {
    return redirect('/auth/forgot-password');
  }
  return { token };

}

export async function action({ request }:  Route.ActionArgs) {
  let formData = await request.formData();
  const url = new URL(request.url);
  const searchParams = url.searchParams;
  const token = searchParams.get('token');
  const password = formData.get('password') as string;
  const confirmPassword = formData.get('confirmPassword') as string;
  if (password !== confirmPassword) {
    return {
      status: 400,
      body: {
        message: 'Passwords do not match',
      }
    };
  }
  if (!password || !token) {
    return {
      status: 400,
      body: {
        message: 'Password and token are required',
      }
    };
  }
  const result = await authClient.passwordReset({
    params: {
      token
    },
    body: {
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

const initialPasswordResetParams = {
  password: '',
  confirmPassword: ''
}

export default function Page() {
  const { token } = useLoaderData();


  const resetPasswordHeader = (
    <>
      <CardTitle>Reset your password</CardTitle>
      <CardDescription>Enter your new password</CardDescription>
    </>
  )

  const resetPasswordFields: CField[] = [
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

  const resetPasswordButtons = (
    <Button type="submit" className="w-full rounded-xl">
      Reset
    </Button>
  )


  return (
    <PublicFormLayout cardHeader={resetPasswordHeader} cardContent={
      <>
        <CForm initialBody={initialPasswordResetParams} fields={resetPasswordFields} buttons={resetPasswordButtons} />
      </>
    } />
  );
}
