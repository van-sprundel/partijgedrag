import { Form, useActionData } from 'react-router';
import { Label } from '~/components/ui/label';
import { Input } from '~/components/ui/input';
import { Button } from '~/components/ui/button';
import React from 'react';
import { useFormHook } from '~/hooks/form-hook';
import type { AuthClientResponseShapes } from '~/clients';
import { Alert, AlertDescription, AlertTitle } from '~/components/ui/alert';
import { AlertCircle } from 'lucide-react';

export type CField = {
  name: string;
  type: 'text' | 'password' | 'email';
  label: string;
  placeholder?: string;
  required?: boolean;
  value?: string | number | string[];
  rightContent?: React.ReactNode;
}

export type CFormProps = {
  initialBody: Record<string, string>;
  fields: CField[];
  buttons?: React.ReactNode;
  method?: 'post' | 'get';
  action?: string;
}


// fix generic type
export default function CForm(props: CFormProps) {
  const [formBody, handleChange] = useFormHook<Record<string, string>>(props.initialBody);

  const actionData = useActionData() as AuthClientResponseShapes | undefined;
  const status = actionData?.status;
  const body = actionData?.body;


  return (
    <>
      <Form method={props.method ?? 'post'} action={props.action} onSubmit={() => console.log("Form submitted")}>
        <div className="flex flex-col gap-6">
          {
            props.fields.map(field => (
              <div className="grid gap-3">
                <Label htmlFor={field.name}>{field.label}</Label>
                <Input
                  id={field.name}
                  type={field.type}
                  name={field.name}
                  placeholder={field.placeholder}
                  value={formBody[field.name]}
                  onChange={handleChange(field.name)}
                  required={field.required}
                />
                {field.rightContent}
              </div>
            ))
          }
          {
            status && status !== 200 &&
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>
                { body?.message ?? 'An error occurred'}
              </AlertDescription>
            </Alert>
          }
          <div className="flex flex-col gap-3">
            {
              props.buttons
            }
          </div>
        </div>
      </Form>
    </>
  )
}