"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuPortal,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { DropdownMenuLink } from "@/components/ui/dropdown-menu-link";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Edit,
  Laptop,
  LogOut,
  Moon,
  MoonStar,
  MoreVertical,
  Plus,
  Rocket,
  Sun,
  Trash,
  User,
} from "lucide-react";
import { useTheme } from "next-themes";
import Image from "next/image";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useLocalStorage } from "usehooks-ts";
import * as z from "zod";
import logoDark from "../assets/logo-dark.svg";
import logoLight from "../assets/logo-light.svg";

const appPasswordFormSchema = z.object({
  appPassword: z.string().min(1, "App password is required"),
});

export default function Home() {
  const { setTheme } = useTheme();
  const [cardCount, setCardCount] = useState(1);
  const [createFeedDialogOpen, setCreateFeedDialogOpen] = useState(false);

  const [appPassword, setAppPassword] = useLocalStorage(
    "atmosfeed.apppassword",
    ""
  );

  const appPasswordForm = useForm<z.infer<typeof appPasswordFormSchema>>({
    resolver: zodResolver(appPasswordFormSchema),
    defaultValues: {
      appPassword: "",
    },
  });

  return (
    <>
      <div className="fixed w-full">
        <header className="container flex justify-between items-center py-6">
          <Image
            src={logoDark}
            alt="Atmosfeed Logo"
            className="h-10 w-auto mr-4 logo-dark"
          />

          <Image
            src={logoLight}
            alt="Atmosfeed Logo"
            className="h-10 w-auto mr-4 logo-light"
          />

          <div className="flex content-center">
            <Dialog open={appPassword === ""}>
              <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                  <DialogTitle>Enter App Password</DialogTitle>
                  <DialogDescription>
                    Atmosfeed need an{" "}
                    <a
                      className="underline"
                      href="https://bsky.app/settings/app-passwords"
                      target="_blank"
                    >
                      app password
                    </a>{" "}
                    to work. It is only stored in your browser and never
                    uploaded to our servers.
                  </DialogDescription>
                </DialogHeader>

                <Form {...appPasswordForm}>
                  <form
                    onSubmit={appPasswordForm.handleSubmit((v) =>
                      setAppPassword(v.appPassword)
                    )}
                    className="space-y-8"
                    id="appPassword"
                  >
                    <FormField
                      control={appPasswordForm.control}
                      name="appPassword"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>App Password</FormLabel>

                          <FormControl>
                            <Input type="password" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </form>
                </Form>

                <DialogFooter>
                  <Button type="submit" form="appPassword">
                    <Plus className="sm:mr-2 h-4 w-4" /> Save
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>

            <Dialog
              onOpenChange={(v) => setCreateFeedDialogOpen(v)}
              open={createFeedDialogOpen}
            >
              <DialogTrigger asChild>
                <Button
                  className="mr-4"
                  onClick={() => setCreateFeedDialogOpen((v) => !v)}
                >
                  <Plus className="sm:mr-2 h-4 w-4" />{" "}
                  <span className="hidden sm:inline">Create Feed</span>
                </Button>
              </DialogTrigger>

              <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                  <DialogTitle>Create Feed</DialogTitle>
                  <DialogDescription>
                    Build a new custom Bluesky feed.
                  </DialogDescription>
                </DialogHeader>

                <div className="grid gap-4 py-4">
                  <div className="grid w-full gap-4">
                    <Label htmlFor="title">Title</Label>
                    <Input id="title" value="Atmosfeed Trending" />
                  </div>

                  <div className="grid w-full gap-4">
                    <Label htmlFor="key">Key</Label>
                    <Input id="key" value="trending" />
                  </div>

                  <div className="grid w-full gap-4">
                    <Label htmlFor="description">Description</Label>
                    <Textarea
                      id="description"
                      placeholder="Most popular trending posts on Bluesky in the last 24h (testing feed)"
                    />
                  </div>
                </div>

                <DialogFooter>
                  <Button
                    onClick={() => {
                      setCardCount((v) => v + 1);
                      setCreateFeedDialogOpen((v) => !v);
                    }}
                  >
                    <Plus className="sm:mr-2 h-4 w-4" /> Create Feed
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>

            <DropdownMenu>
              <DropdownMenuTrigger>
                <Avatar>
                  <AvatarImage
                    src="https://github.com/pojntfx.png"
                    alt="@pojntfx"
                  />
                  <AvatarFallback>AV</AvatarFallback>
                </Avatar>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuLabel>My Account</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuLink
                  href="https://bsky.app/profile/felicitas.pojtinger.com"
                  target="_blank"
                >
                  <User className="mr-2 h-4 w-4" /> Profile
                </DropdownMenuLink>
                <DropdownMenuItem onClick={() => setAppPassword("")}>
                  <LogOut className="mr-2 h-4 w-4" /> Logout
                </DropdownMenuItem>

                <DropdownMenuLabel>Settings</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger>
                    <MoonStar className="mr-2 h-4 w-4" />
                    <span>Theme</span>
                  </DropdownMenuSubTrigger>
                  <DropdownMenuPortal>
                    <DropdownMenuSubContent>
                      <DropdownMenuItem onClick={() => setTheme("light")}>
                        <Sun className="mr-2 h-4 w-4" /> Light
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => setTheme("dark")}>
                        <Moon className="mr-2 h-4 w-4" /> Dark
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => setTheme("system")}>
                        <Laptop className="mr-2 h-4 w-4" /> System
                      </DropdownMenuItem>
                    </DropdownMenuSubContent>
                  </DropdownMenuPortal>
                </DropdownMenuSub>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>

        <div className="gradient-blur">
          <div></div>
          <div></div>
          <div></div>
          <div></div>
          <div></div>
          <div></div>
        </div>

        <div className="gradient-blur-bottom">
          <div></div>
          <div></div>
          <div></div>
          <div></div>
          <div></div>
          <div></div>
        </div>
      </div>

      <div className="content">
        <main className="flex-grow flex flex-col justify-center items-center gap-2 container">
          {[...Array(cardCount).keys()].map((_, i) => (
            <Card
              className="w-full max-w-2xl flex items-center justify-between"
              key={i}
            >
              <CardHeader>
                <div className="text-2xl font-semibold leading-none tracking-tight flex items-center justify-between">
                  <div>Atmosfeed Trending {i + 1}</div>

                  <DropdownMenu>
                    <DropdownMenuTrigger asChild className="sm:hidden">
                      <Button variant="ghost" size="icon">
                        <MoreVertical />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent>
                      <DropdownMenuGroup>
                        <DropdownMenuItem>
                          <Edit className="mr-2 h-4 w-4" /> Edit
                        </DropdownMenuItem>

                        <DropdownMenuItem>
                          <Rocket className="mr-2 h-4 w-4" /> Publish
                        </DropdownMenuItem>

                        <DropdownMenuItem>
                          <Trash className="mr-2 h-4 w-4" /> Delete
                        </DropdownMenuItem>
                      </DropdownMenuGroup>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
                <CardDescription>
                  <code>trending-{i + 1}</code>
                </CardDescription>
                <CardDescription>
                  Most popular trending posts on Bluesky in the last 24h
                  (testing feed)
                </CardDescription>
              </CardHeader>

              <CardFooter className="py-0 pr-4 hidden sm:flex">
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

                      <DropdownMenuItem>
                        <Rocket className="mr-2 h-4 w-4" /> Publish
                      </DropdownMenuItem>

                      <DropdownMenuItem>
                        <Trash className="mr-2 h-4 w-4" /> Delete
                      </DropdownMenuItem>
                    </DropdownMenuGroup>
                  </DropdownMenuContent>
                </DropdownMenu>
              </CardFooter>
            </Card>
          ))}
        </main>
      </div>

      <div className="fixed bottom-0 w-full">
        <footer className="flex justify-between items-center py-6 container">
          <a
            href="https://github.com/pojntfx/atmosfeed"
            target="_blank"
            className="hover:underline whitespace-nowrap mr-4"
          >
            © 2023 Felicitas Pojtinger
          </a>

          <a
            href="https://felicitas.pojtinger.com/imprint"
            target="_blank"
            className="hover:underline"
          >
            Imprint
          </a>
        </footer>
      </div>
    </>
  );
}
