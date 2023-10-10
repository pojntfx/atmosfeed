"use client";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
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
import { FeedCard } from "@/components/ui/feed-card";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PrivacyPolicy } from "@/components/ui/privacy-policy";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/components/ui/use-toast";
import { useAPI } from "@/hooks/api";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Database,
  DownloadCloud,
  Laptop,
  Loader,
  LogIn,
  LogOut,
  Moon,
  MoonStar,
  Plus,
  Sun,
  TrashIcon,
  User,
} from "lucide-react";
import { useTheme } from "next-themes";
import Image from "next/image";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useLocalStorage } from "usehooks-ts";
import * as z from "zod";
import logoDark from "../assets/logo-dark.png";
import logoLight from "../assets/logo-light.png";

const setupFormSchema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "App password is required"),

  service: z.string().min(1, "Service is required"),
  atmosfeedAPI: z.string().min(1, "Atmosfeed API is required"),
  feedGeneratorDID: z.string().min(1, "Feed generator DID is required"),

  acceptedPrivacyPolicy: z.literal<boolean>(true),
});

const createFeedSchema = z.object({
  rkey: z.string().min(1, "Resource key is required"),

  classifier: z.instanceof(File, {
    message: "Classifier is required",
  }),
});

export default function Home() {
  const { setTheme } = useTheme();
  const [loginDialogOpen, setLoginDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [privacyPolicyDialogOpen, setPrivacyPolicyDialogOpen] = useState(false);

  const [username, setUsername] = useLocalStorage("atmosfeed.username", "");
  const [password, setPassword] = useLocalStorage("atmosfeed.password", "");

  const [service, setService] = useLocalStorage(
    "atmosfeed.service",
    process.env.ATMOSFEED_SERVICE_DEFAULT || "https://bsky.social"
  );
  const [atmosfeedAPI, setAtmosfeedAPI] = useLocalStorage(
    "atmosfeed.atmosfeedURL",
    process.env.ATMOSFEED_API_DEFAULT || "https://manager.atmosfeed.p8.lu"
  );
  const [feedGeneratorDID, setFeedGeneratorDID] = useLocalStorage(
    "atmosfeed.feedGeneratorDID",
    process.env.ATMOSFEED_FEED_GENERATOR_DID_DEFAULT ||
      "did:web:atmosfeed-feeds.serveo.net"
  );

  const setupForm = useForm<z.infer<typeof setupFormSchema>>({
    resolver: zodResolver(setupFormSchema),
    defaultValues: {
      username,
      password,

      service,
      atmosfeedAPI,
      feedGeneratorDID,

      acceptedPrivacyPolicy: false,
    },
  });

  const createFeedForm = useForm<z.infer<typeof createFeedSchema>>({
    resolver: zodResolver(createFeedSchema),
    defaultValues: {},
  });

  const [createFeedDialogOpen, setCreateFeedDialogOpen] = useState(false);

  const { toast } = useToast();

  const {
    avatar,
    did,
    signedIn,

    unpublishedFeeds,
    publishedFeeds,

    applyFeed,

    deleteData,

    loading,
    logout,
  } = useAPI(
    username,
    password,
    service,
    atmosfeedAPI,
    () => setPassword(""),
    (err, loggedOut) =>
      loggedOut
        ? toast({
            title: "You Have Been Logged Out",
            description: `Authentication with Bluesky failed and you have been logged out. The error is: "${err?.message}"`,
          })
        : toast({
            title: "An Error Occured",
            description: `An error could not be handled. The error is: "${err?.message}"`,
          })
  );

  return (
    <>
      <div className="fixed w-full">
        <header className="container flex justify-between items-center py-6">
          {signedIn && (
            <Image
              src={logoDark}
              alt="Atmosfeed Logo"
              className="h-10 w-auto mr-4 logo-dark"
            />
          )}

          {signedIn && (
            <Image
              src={logoLight}
              alt="Atmosfeed Logo"
              className="h-10 w-auto mr-4 logo-light"
            />
          )}

          {signedIn && (
            <DropdownMenu>
              <DropdownMenuTrigger>
                <Avatar>
                  <AvatarImage src={avatar} alt={"Avatar of " + username} />
                  <AvatarFallback>AV</AvatarFallback>
                </Avatar>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuLabel>My Account</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuLink
                  href={`https://bsky.app/profile/${username}`}
                  target="_blank"
                >
                  <User className="mr-2 h-4 w-4" /> Profile
                </DropdownMenuLink>
                <DropdownMenuItem onClick={() => logout()}>
                  <LogOut className="mr-2 h-4 w-4" /> Logout
                </DropdownMenuItem>

                <DropdownMenuSub>
                  <DropdownMenuSubTrigger>
                    <Database className="mr-2 h-4 w-4" />
                    <span className="mr-2">Your Data</span>
                  </DropdownMenuSubTrigger>

                  <DropdownMenuPortal>
                    <DropdownMenuSubContent>
                      <DropdownMenuItem
                        onClick={() => {
                          const data = {
                            did,
                            service,
                          };

                          const blob = new Blob(
                            [JSON.stringify(data, null, 2)],
                            { type: "application/json" }
                          );
                          const url = URL.createObjectURL(blob);
                          const a = document.createElement("a");

                          a.href = url;
                          a.download = "atmosfeed.json";
                          a.click();

                          URL.revokeObjectURL(url);

                          toast({
                            title: "Data Downloaded Successfully",
                            description:
                              "Your data has successfully been downloaded to your system.",
                          });
                        }}
                      >
                        <DownloadCloud className="mr-2 h-4 w-4" />
                        <span>Download your Data</span>
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={() => setDeleteDialogOpen((v) => !v)}
                      >
                        <TrashIcon className="mr-2 h-4 w-4" />
                        <span>Delete your Data</span>
                      </DropdownMenuItem>
                    </DropdownMenuSubContent>
                  </DropdownMenuPortal>
                </DropdownMenuSub>

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
          )}
        </header>

        {signedIn && (
          <div className="gradient-blur">
            <div></div>
            <div></div>
            <div></div>
            <div></div>
            <div></div>
            <div></div>
          </div>
        )}

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
          {signedIn ? (
            <>
              <div className="w-full max-w-2xl flex flex-col gap-2">
                <div className="flex justify-between items-center gap-2 mb-2">
                  <h2 className="text-xl font-medium">Unpublished Feeds</h2>

                  <Button onClick={() => setCreateFeedDialogOpen(true)}>
                    <Plus className="sm:mr-2 h-4 w-4" />{" "}
                    <span className="hidden sm:inline">Create Feed</span>
                  </Button>
                </div>

                {unpublishedFeeds.length > 0
                  ? unpublishedFeeds.map((feed, i) => (
                      <FeedCard feed={feed} key={i} />
                    ))
                  : "No feeds yet"}
              </div>

              {publishedFeeds.length > 0 && (
                <div className="w-full max-w-2xl flex flex-col gap-2">
                  <div className="flex justify-between items-center gap-2 my-2">
                    <h2 className="text-xl font-medium">Published Feeds</h2>
                  </div>

                  {publishedFeeds.map((feed, i) => (
                    <FeedCard feed={feed} key={i} />
                  ))}
                </div>
              )}
            </>
          ) : (
            <>
              <Image
                src={logoDark}
                alt="Atmosfeed Logo"
                className="h-20 w-auto logo-dark"
              />

              <Image
                src={logoLight}
                alt="Atmosfeed Logo"
                className="h-20 w-auto logo-light"
              />

              <h2 className="text-2xl mt-4 my-5 text-center">
                Create custom Bluesky feeds with WebAssembly and Scale.
              </h2>

              <Button
                disabled={loading}
                onClick={() => setLoginDialogOpen(true)}
                className="mb-10"
              >
                {loading ? (
                  <Loader className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <LogIn className="mr-2 h-4 w-4" />
                )}{" "}
                Login with Bluesky
              </Button>
            </>
          )}
        </main>
      </div>

      <div className="fixed bottom-0 w-full overflow-x-auto">
        <footer className="flex justify-between items-center py-6 container pr-0">
          <a
            href="https://github.com/pojntfx/atmosfeed"
            target="_blank"
            className="hover:underline whitespace-nowrap mr-4"
          >
            © 2023 Felicitas Pojtinger
          </a>

          <div className="flex h-5 items-center space-x-4 text-sm pr-8">
            <Button
              variant="link"
              className="p-0 h-auto font-normal"
              onClick={() => setPrivacyPolicyDialogOpen((v) => !v)}
            >
              Privacy
            </Button>

            <Separator orientation="vertical" />

            <a
              href="https://felicitas.pojtinger.com/imprint"
              target="_blank"
              className="hover:underline"
            >
              Imprint
            </a>
          </div>
        </footer>
      </div>

      <Dialog
        onOpenChange={(v) => setLoginDialogOpen(v)}
        open={loginDialogOpen}
      >
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <Image
              src={logoDark}
              alt="Atmosfeed Logo"
              className="h-10 object-contain logo-dark"
            />

            <Image
              src={logoLight}
              alt="Atmosfeed Logo"
              className="h-10 object-contain logo-light"
            />

            <DialogTitle className="pt-4">Login</DialogTitle>
            <DialogDescription>
              Atmosfeed needs access to your Bluesky account in order to delete
              posts on your behalf.
            </DialogDescription>
          </DialogHeader>

          <Form {...setupForm}>
            <form
              onSubmit={setupForm.handleSubmit((v) => {
                setUsername(v.username.replace(/^@/, ""));
                setPassword(v.password);

                setService(v.service);
                setAtmosfeedAPI(v.atmosfeedAPI);
                setFeedGeneratorDID(v.feedGeneratorDID);

                setLoginDialogOpen(false);
              })}
              className="space-y-4"
              id="setup"
            >
              <FormField
                control={setupForm.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username</FormLabel>

                    <FormControl>
                      <Input type="text" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={setupForm.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>

                    <FormDescription>
                      You can use an{" "}
                      <a
                        className="underline"
                        href="https://bsky.app/settings/app-passwords"
                        target="_blank"
                      >
                        app password
                      </a>
                      . It is only stored in this browser and never uploaded to
                      our servers.
                    </FormDescription>

                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={setupForm.control}
                name="acceptedPrivacyPolicy"
                render={({ field }) => {
                  const { value, onChange, ...rest } = field;

                  return (
                    <FormItem className="items-top flex space-x-2 space-y-0 items-center">
                      <FormControl>
                        <Checkbox
                          checked={value}
                          onCheckedChange={onChange}
                          {...rest}
                        />
                      </FormControl>

                      <div className="grid gap-1.5 leading-none">
                        <FormLabel className="text-sm font-medium leading-none">
                          I have read and agree to the{" "}
                          <Button
                            variant="link"
                            className="p-0 underline h-auto font-normal"
                            onClick={() =>
                              setPrivacyPolicyDialogOpen((v) => !v)
                            }
                          >
                            privacy policy
                          </Button>
                        </FormLabel>
                      </div>
                    </FormItem>
                  );
                }}
              />

              <Accordion type="single" collapsible>
                <AccordionItem value="item-1">
                  <AccordionTrigger>Advanced</AccordionTrigger>
                  <AccordionContent>
                    <FormField
                      control={setupForm.control}
                      name="service"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Service</FormLabel>

                          <FormDescription>
                            The Bluesky service your account is hosted on; most
                            users don&apos;t need to change this.
                          </FormDescription>

                          <FormControl>
                            <Input type="text" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={setupForm.control}
                      name="atmosfeedAPI"
                      render={({ field }) => (
                        <FormItem className="mt-4">
                          <FormLabel>Atmosfeed API</FormLabel>

                          <FormDescription>
                            The URL that Atmosfeed&apos;s API is hosted on; most
                            users don&apos;t need to change this.
                          </FormDescription>

                          <FormControl>
                            <Input type="text" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={setupForm.control}
                      name="feedGeneratorDID"
                      render={({ field }) => (
                        <FormItem className="mt-4">
                          <FormLabel>Feed Generator DID</FormLabel>

                          <FormDescription>
                            The DID that the feed generator is reachable under,
                            typically the hostname of the publically reachable
                            URL; most users don&apos;t need to change this.
                          </FormDescription>

                          <FormControl>
                            <Input type="text" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </AccordionContent>
                </AccordionItem>
              </Accordion>
            </form>
          </Form>

          <DialogFooter>
            <Button type="submit" form="setup">
              Next
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        onOpenChange={(v) => setCreateFeedDialogOpen(v)}
        open={createFeedDialogOpen}
      >
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Create Feed</DialogTitle>
          </DialogHeader>

          <Form {...createFeedForm}>
            <form
              onSubmit={createFeedForm.handleSubmit(async (v) => {
                await applyFeed(v.rkey, v.classifier);

                setCreateFeedDialogOpen(false);

                toast({
                  title: "Feed Created Successfullyy",
                  description:
                    "Your classifier has been uploaded successfully.",
                });
              })}
              className="space-y-4"
              id="create-feed"
            >
              <FormField
                control={createFeedForm.control}
                name="rkey"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Resource Key</FormLabel>
                    <FormDescription>
                      Machine-readable key for the feed.
                    </FormDescription>

                    <FormControl>
                      <Input type="text" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={createFeedForm.control}
                name="classifier"
                render={({ field: { value, onChange, ...rest } }) => (
                  <FormItem>
                    <FormLabel>Classifier</FormLabel>
                    <FormDescription>
                      Exported Scale function (.scale file) to use as the
                      classifier for the feed.
                    </FormDescription>

                    <FormControl>
                      <Input
                        type="file"
                        placeholder="classifier.scale"
                        accept="application/octet-stream"
                        onChange={(event) =>
                          onChange(event.target.files && event.target.files[0])
                        }
                        {...rest}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </form>
          </Form>

          <DialogFooter>
            <Button type="submit" form="create-feed" disabled={loading}>
              {loading ? (
                <Loader className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Plus className="mr-2 h-4 w-4" />
              )}{" "}
              Create Feed
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        onOpenChange={(v) => setDeleteDialogOpen(v)}
        open={deleteDialogOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete your Atmosfeed account and remove
              your data from our servers. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={async () => {
                await deleteData();

                toast({
                  title: "Data Deleted Successfullyy",
                  description:
                    "Your data has successfully been deleted from our servers and you have been logged out.",
                });
              }}
            >
              Delete Your Data
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Dialog
        onOpenChange={(v) => setPrivacyPolicyDialogOpen(v)}
        open={privacyPolicyDialogOpen}
      >
        <DialogContent className="max-w-[720px] h-[720px] max-h-screen">
          <DialogHeader>
            <DialogTitle>Privacy Policy</DialogTitle>
          </DialogHeader>

          <ScrollArea className="privacy-policy">
            <PrivacyPolicy />
          </ScrollArea>
        </DialogContent>
      </Dialog>
    </>
  );
}
