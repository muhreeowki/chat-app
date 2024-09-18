import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export default function ChatMessage(props: {
  sender: string;
  payload: string;
  datetime: string;
}) {
  return (
    <div className="grid gap-2">
      <div className="flex items-center gap-4">
        <Avatar className="hidden h-9 w-9 sm:flex">
          <AvatarImage src="../../public/avatars/01.png" alt="Avatar" />
          <AvatarFallback>OM</AvatarFallback>
        </Avatar>
        <div className="grid gap-1">
          <p className="text-sm font-medium leading-none">{props.sender}</p>
          <p className="text-sm text-muted-foreground">{props.datetime}</p>
        </div>
      </div>
      <div className="font-medium">{props.payload}</div>
    </div>
  );
}
