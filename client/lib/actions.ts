"use server";

import { cookies } from "next/headers";
import { Message, User, UserReqData } from "@/lib/types";

export async function getMessages(): Promise<Message[]> {
  const usr: User = await getUserData();

  if (usr.token !== "") {
    const messages: Message[] = await fetch("http://chatserver:8080/messages", {
      cache: "no-store",
      headers: {
        Authorization: usr.token,
      },
    })
      .then((value) => value)
      .then((data) => data.json())
      .catch((e) => {
        console.error(e);
        return [];
      });
    return messages;
  }
  return [];
}

export async function getUserData(): Promise<User> {
  const data = cookies().get("userData")?.value || "";
  if (data === "") {
    return { id: "", username: "", token: "" };
  }
  const usr: User = JSON.parse(data);
  return usr;
}

export async function login(
  username: string,
  password: string,
): Promise<boolean> {
  const usr: User = await fetch("http://chatserver:8080/login", {
    method: "POST",
    body: JSON.stringify({
      username: username,
      password: password,
    }),
  })
    .then((res) => res.json())
    .then((data) => data)
    .catch((e) => {
      return false;
    });

  if (!usr.token) {
    return false;
  }

  cookies().set("userData", JSON.stringify(usr), { httpOnly: true });
  return true;
}

export async function signup(
  username: string,
  password: string,
): Promise<boolean> {
  const usr: User = await fetch("http://chatserver:8080/signup", {
    method: "POST",
    body: JSON.stringify({
      username: username,
      password: password,
    }),
  })
    .then((res) => res.json())
    .then((data) => data)
    .catch((e) => {
      return false;
    });

  if (!usr.token) {
    return false;
  }

  cookies().set("userData", JSON.stringify(usr), { httpOnly: true });
  return true;
}
