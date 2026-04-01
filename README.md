```
 ██████╗██╗   ██╗██████╗ ███████╗██████╗ ███████╗██████╗  █████╗  ██████╗███████╗
██╔════╝╚██╗ ██╔╝██╔══██╗██╔════╝██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔════╝██╔════╝
██║      ╚████╔╝ ██████╔╝█████╗  ██████╔╝███████╗██████╔╝███████║██║     █████╗
██║       ╚██╔╝  ██╔══██╗██╔══╝  ██╔══██╗╚════██║██╔═══╝ ██╔══██║██║     ██╔══╝
╚██████╗   ██║   ██████╔╝███████╗██║  ██║███████║██║     ██║  ██║╚██████╗███████╗
 ╚═════╝   ╚═╝   ╚═════╝ ╚══════╝╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝ ╚═════╝╚══════╝
```

### Cyberspace CLI Client Prototype

At present, this client only has 3 real functions: Browse the feed, check your notifications, look at the replies to individual posts & write posts of your own.

To download it, you need to have Go installed. So long as that is given, you can simply clone the Git repo onto your machine (or download it via github), then while inside the project directory, execute the following commands: 

```go
go build -o cyberspace-client .
./cyberspace-client
```

Btw, it you're a programmer, please don't look at the contents of this too closely, especially main.go. You don't want to see what's going on in there. 

The client has the following commands:

- feed: Load 5 posts from the cyberspace feed. Each use loads the next 5 after the last one.
- write: Opens your default text editor (or if you have non, vim) and lets you write a post. Be aware that it might fail to post, so don't invest too much effort into it without copying the contents elsewhere before saving and closing the editor. After closing the editor, you'll have a chance to choose topics for the post.
- post <post_id>: Shows a post and all its replies. This command requires an argument, the post id shown in the top-line of each post.
- notifications: Get the last 10 notifications, and the 10 more with each additional use. If the notification is for a new post, it will also show the id of that post so you can open it with the 'post' command. 
- note <note_id>: Lets you open an existing note from your journal, edit it, and save the edited version.

### Warnings:
When the program closes, including via ctrl+c, it will reset your terminal color to the default. 