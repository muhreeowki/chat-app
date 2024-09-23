"use server";

import { cookies } from "next/headers";
import { Message, User, UserReqData } from "@/lib/types";

export async function GetMessages(): Promise<Message[]> {
  const token = cookies().get("authToken")?.value || "";

  console.log(token);
  if (token !== "") {
    const messages: Message[] = await fetch("http://localhost:8080/messages", {
      cache: "no-store",
      headers: {
        Authorization: token,
      },
    })
      .then((value) => value)
      .then((data) => data.json())
      .catch((e) => console.error(e));
    console.log(messages);
    return messages;
  }
  return [];
}

export async function login(
  username: string,
  password: string,
): Promise<boolean> {
  const usr: User = await fetch("http://localhost:8080/login", {
    method: "POST",
    body: JSON.stringify({
      username: username,
      password: password,
    }),
  })
    .then((res) => res.json())
    .then((data) => data)
    .catch((e) => {
      console.error(e);
      return false;
    });

  if (!usr.token) {
    console.error(usr);
    return false;
  }

  cookies().set("authToken", usr.token, { httpOnly: true });
  return true;
}

export async function signup(
  username: string,
  password: string,
): Promise<boolean> {
  const usr: User = await fetch("http://localhost:8080/signup", {
    method: "POST",
    body: JSON.stringify({
      username: username,
      password: password,
    }),
  })
    .then((res) => res.json())
    .then((data) => data)
    .catch((e) => {
      console.error(e);
      return false;
    });

  if (!usr.token) {
    console.error(usr);
    return false;
  }

  cookies().set("authToken", usr.token, { httpOnly: true });
  return true;
}
