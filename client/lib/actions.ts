"use server";

import { cookies } from "next/headers";
import { Message, User, UserReqData } from "@/lib/types";
import axios from "axios";

export async function getMessages(): Promise<Message[]> {
  const usr: User = await getUserData();

  if (usr.token !== "") {
    const res = await axios.get("http://chatserver:8080/messages", {
      headers: {
        Authorization: usr.token,
      },
    });
    if (res.status === 200) {
      return res.data;
    } else {
      return [];
    }
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
  const res = await axios.post("http://chatserver:8080/login", {
    username: username,
    password: password,
  });

  if (res.status !== 200) {
    return false;
  }

  const usr: User = res.data;

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
  const res = await axios.post("http://chatserver:8080/signup", {
    username: username,
    password: password,
  });

  if (res.status !== 201) {
    return false;
  }

  const usr: User = res.data;

  if (!usr.token) {
    return false;
  }

  cookies().set("userData", JSON.stringify(usr), { httpOnly: true });
  return true;
}
