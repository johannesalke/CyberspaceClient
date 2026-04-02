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


### Getting Started

To download it, you need to have Go installed. So long as that is given, you can simply clone the Git repo onto your machine (or download it via github), then while inside the project directory, execute the following commands: 

```go
go build -o cyberspace-client .
./cyberspace-client
```

Btw, it you're a programmer, please don't look at the contents of this too closely, especially main.go. You don't want to see what's going on in there. 

### How to use 

Client commands consist of a verb and a noun. 

- `view feed (*optional_arg*)`: Load 5 posts from the feed, starting at the newest. Every time the command is used, 5 more are loaded starting from where the previous iteration stopped. Use the optional argument 'new' to load posts made since you started the client without losing the marker of the basic command. Use 'reset' to start over entirely.
- `view post <post_id>`: This command shows the post specified by the id argument, plus the first 20 comments.
- `view notifications (*optional_arg*)`: Load 10 notifications. If the notification is for a post or reply, you can use the shown id to open that post. Supports the same optional arguments as 'view feed'
- `view notes`: Loads 10 notes from your journal.
- `write post`: Opens your default text editor (or if you have non, nano (use ctrl+s, ctrl+x to exit)) and lets you write a post. Be aware that it might fail to post, so don't invest too much effort into it without copying the contents elsewhere before saving and closing the editor. After closing the editor, you'll have a chance to choose topics for the post.
- `write note`: Same as 'write post', but your writing is put in your journal instead.
- `edit note <note_id`: Opens a note in your default text editor (if none, nano) and lets you edit it.
- `post <note_id>`: Posts a note to the feed, making it visible to other users. 
- `edit config`: This lets you edit the client's config file. If you set 'stay logged in' to true, the client will save your refresh token and you will remain logged in across sessions. The config file should be in your .config or appdata folder, depending on whether you use linux or windows. 






### Warnings:
When the program closes, including via ctrl+c, it will reset your terminal color to the default. 