"use client";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useState } from "react";
import { LoginFormSchema } from "@/lib/types";
import { signup, login } from "@/lib/actions";
import { useRouter } from "next/navigation";

export default function LoginForm() {
  const [newUser, setNewUser] = useState(false);

  const router = useRouter();

  const form = useForm<z.infer<typeof LoginFormSchema>>({
    resolver: zodResolver(LoginFormSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  async function handler(usrData: z.infer<typeof LoginFormSchema>) {
    let ok;
    if (newUser) {
      ok = await signup(usrData.username, usrData.password);
    } else {
      ok = await login(usrData.username, usrData.password);
    }
    if (ok) {
      router.push("/");
    }
  }

  return (
    <div className="grid items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-2xl">
            {newUser ? "Create an Account" : "Login"}
          </CardTitle>
          <CardDescription>
            Enter your username and password to get started
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(handler)} className="space-y-6">
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username</FormLabel>
                    <FormControl>
                      <Input placeholder="bob" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" className="w-full">
                {newUser ? "Sign Up" : "Login"}
              </Button>
              {newUser ? (
                <p
                  className="w-full text-muted-foreground text-sm cursor-pointer"
                  onClick={() => setNewUser(false)}
                >
                  Already have an account?{" "}
                  <span className="text-blue-600 hover:underline hover:text-blue-800">
                    Login
                  </span>
                </p>
              ) : (
                <p
                  className="w-full text-muted-foreground text-sm cursor-pointer"
                  onClick={() => setNewUser(true)}
                >
                  Don't have an account?{" "}
                  <span className="text-blue-600 hover:underline hover:text-blue-800">
                    Create a New Account
                  </span>
                </p>
              )}
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
