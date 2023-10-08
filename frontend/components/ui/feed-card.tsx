import { IFeed } from "@/api/models";
import {
  Edit,
  FileSignature,
  MoreVertical,
  PlaneLanding,
  Trash,
} from "lucide-react";
import { Button } from "./button";
import { Card, CardDescription, CardFooter, CardHeader } from "./card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./dropdown-menu";

export const FeedCard: React.FC<{ feed: IFeed }> = ({
  feed,
  ...otherProps
}) => (
  <Card className="flex items-center justify-between" {...otherProps}>
    <CardHeader>
      <div className="text-2xl font-semibold leading-none tracking-tight flex items-center justify-between">
        {feed.title && <div>{feed.title}</div>}
      </div>
      <CardDescription className={feed.title ? "" : "title-description"}>
        <code>{feed.rkey}</code>
      </CardDescription>

      {feed.description && (
        <CardDescription>{feed.description}</CardDescription>
      )}
    </CardHeader>

    <CardFooter className="py-0 pr-4 gap-2">
      {!feed.title && (
        <Button variant="secondary" className="hidden sm:flex">
          <FileSignature className="mr-2 h-4 w-4" /> Finalize
        </Button>
      )}

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <MoreVertical />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuGroup>
            <DropdownMenuItem>
              <Edit className="mr-2 h-4 w-4" /> Edit
            </DropdownMenuItem>

            {feed.title ? (
              <DropdownMenuItem>
                <PlaneLanding className="mr-2 h-4 w-4" /> Unpublish
              </DropdownMenuItem>
            ) : (
              <DropdownMenuItem className="sm:hidden">
                <FileSignature className="mr-2 h-4 w-4" /> Finalize
              </DropdownMenuItem>
            )}

            <DropdownMenuItem>
              <Trash className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </CardFooter>
  </Card>
);
