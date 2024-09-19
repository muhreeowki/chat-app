"use server";

import { Message } from "@/lib/types";

export async function GetMessages(): Promise<Message[]> {
  const messages: Message[] = await fetch("http://localhost:8080/messages", {
    cache: "no-store",
  }).then((value) => value.json());

  console.log(messages);
  return messages;
}
