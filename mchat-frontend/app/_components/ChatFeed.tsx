"use client";

import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CornerDownLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import ChatMessage from "./ChatMessage";
import ChatBox, { formSchema } from "./ChatBox";
import { GetMessages, Message } from "@/actions/getMessages";
import { useState } from "react";
import { useToast } from "@/hooks/use-toast";
import { z } from "zod";

export default function ChatFeed() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [connected, setConnected] = useState(false);
  const [ws, setWS] = useState<WebSocket | undefined>(undefined);
  const { toast } = useToast();

  function handleSendMessage(data: z.infer<typeof formSchema>) {
    if (connected) {
      ws?.send(
        JSON.stringify({
          Payload: data.message,
          Sender: "Michele",
        }),
      );
    } else {
      toast({
        title: "Not Connected",
        description: "You are not connected to the chat server.",
        variant: "destructive",
      });
    }
  }

  async function handleConnect() {
    if (!connected || ws == undefined) {
      const newWS = new WebSocket("ws://localhost:4000");
      newWS.onclose = (event) => {
        console.log(event);
        setConnected(false);
      };
      newWS.onmessage = async (event) => {
        setMessages(await GetMessages());
        console.log(event.data);
      };
      setWS(newWS);
      setConnected(true);
    }
    setMessages(await GetMessages());
  }

  return (
    <Card className="max-w-screen-md w-full">
      <CardHeader>
        <CardTitle>Mchat</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col-reverse max-h-[500px] overflow-scroll">
        {connected ? (
          <div className="grid gap-8">
            {messages.map((msg, i) => (
              <ChatMessage
                key={i}
                payload={msg.Payload}
                sender={msg.Sender}
                datetime={msg.Datetime}
              />
            ))}
          </div>
        ) : (
          <div className="items-center rounded-lg bg-background w-full">
            <Button
              type="button"
              onClick={handleConnect}
              size="sm"
              className="ml-auto gap-1.5"
            >
              Connect to Send Messages
              <CornerDownLeft className="size-3.5" />
            </Button>
          </div>
        )}
      </CardContent>
      {connected && (
        <CardFooter className="grid">
          <ChatBox handleSendMessage={handleSendMessage} />
        </CardFooter>
      )}
    </Card>
  );
}
