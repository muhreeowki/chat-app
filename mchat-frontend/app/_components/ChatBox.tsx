"use client";

import { CornerDownLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useState } from "react";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { SubmitHandler, useForm } from "react-hook-form";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";

export const formSchema = z.object({
  message: z.string().min(1).max(180),
});

export default function ChatBox(props: {
  handleSendMessage: SubmitHandler<{ message: string }>;
}) {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      message: "",
    },
  });

  return (
    <Form {...form}>
      <form
        className="relative overflow-hidden rounded-lg border bg-background focus-within:ring-1 focus-within:ring-ring"
        onSubmit={form.handleSubmit(props.handleSendMessage)}
      >
        <FormField
          control={form.control}
          name="message"
          render={({ field }: any) => (
            <FormItem className="p-3">
              <FormControl>
                <Textarea
                  {...field}
                  placeholder="Type your message here..."
                  className="min-h-13 resize-none border-0 shadow-none focus-visible:ring-0"
                />
              </FormControl>
              <div className="flex items-center p-3 pt-0">
                <FormMessage />
                <Button type="submit" size="sm" className="ml-auto gap-1.5">
                  Send Message
                  <CornerDownLeft className="size-3.5" />
                </Button>
              </div>
            </FormItem>
          )}
        />
      </form>
    </Form>
  );
}
