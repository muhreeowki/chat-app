import z from "zod";

export interface Message {
  payload: string;
  sender: string;
  datetime: string;
}

export interface User {
  id: string;
  username: string;
  token: string;
}

export interface UserReqData {
  username: string;
  password: string;
}

export const LoginFormSchema = z.object({
  username: z.string().min(2).max(50),
  password: z.string().min(4).max(50),
});
