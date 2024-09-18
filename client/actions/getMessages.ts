"use server";

export interface Message {
  Payload: string;
  Sender: string;
  Datetime: string;
}

export async function GetMessages(): Promise<Message[]> {
  const messages: Message[] = await fetch("http://localhost:8080/", {
    cache: "no-store",
  }).then((value) => value.json());

  console.log(messages);
  return messages;
}
