```
 ██████╗██╗   ██╗██████╗ ███████╗██████╗ ███████╗██████╗  █████╗  ██████╗███████╗
██╔════╝╚██╗ ██╔╝██╔══██╗██╔════╝██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔════╝██╔════╝
██║      ╚████╔╝ ██████╔╝█████╗  ██████╔╝███████╗██████╔╝███████║██║     █████╗
██║       ╚██╔╝  ██╔══██╗██╔══╝  ██╔══██╗╚════██║██╔═══╝ ██╔══██║██║     ██╔══╝
╚██████╗   ██║   ██████╔╝███████╗██║  ██║███████║██║     ██║  ██║╚██████╗███████╗
 ╚═════╝   ╚═╝   ╚═════╝ ╚══════╝╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝ ╚═════╝╚══════╝
```

### Cyberspace CLI Client Prototype

At present, this client has solid basic functions and presentation, but is definitely not yet finished.
You can check your feed and notifications, make posts as well as write, edit and publish notes.
You can not yet reply to people, check their profiles, or search for specific topics. 


### Getting Started

This assumes you already have a Go itself installed. If you do not, refer to this page [this page](https://golang.org/doc/install) before returning. 

So long as you have the client installed, you can simply clone the Git (`git clone github.com/johannesalke/cyberspacecli`) repo onto your machine (or download it via github), then while inside the project directory execute the following commands: 

```go
go build -o cyberspacecli .
./cyberspacecli
```
The first comiles the program, the second executes it. (To exit, use the command 'exit' or press the keys ctrl+C.)

If you like what you see and want to have the tool available anywhere on you machine (as opposed to just inside this folder) you can also install it as a command line tool. You can then use it as easily as you would git or vim. For that, see the following instructions:

```
go install github.com/johannesalke/cyberspacecli
```

### How to use 

Client commands consist of a verb and a noun. 

- `view feed (optional_arg)`: Load 10 posts from the feed, starting at the newest. Every time the command is used, 10 more are loaded starting from where the previous iteration stopped. In the feed, posts are truncated at 1000 characters. To see the whole post, use the 'view post' command. 
Use the optional argument 'new' to load posts made since you started the client without losing the marker of the basic command. Use 'reset' to start over entirely. (Sometimes it may show less than 10 because NSFW posts are currently filtered by default. In the future, this will be handled based on user settings.)
- `view post <post_id>`: This command shows the post specified by the id argument, plus the first 20 comments.
- `view notifications (optional_arg)`: Load 10 notifications. If the notification is for a post or reply, you can use the shown id to open that post. Supports the same optional arguments as 'view feed'
- `view notes`: Loads 10 notes from your journal.
- `write post`: Opens your default text editor (or if you have non, nano (use ctrl+s, ctrl+x to exit)) and lets you write a post. Be aware that it might fail to post, so don't invest too much effort into it without copying the contents elsewhere before saving and closing the editor. After closing the editor, you'll have a chance to choose topics for the post.
- `write note`: Same as 'write post', but your writing is put in your journal instead.
- `edit note <note_id`: Opens a note in your default text editor (if none, nano) and lets you edit it.
- `post <note_id>`: Posts a note to the feed, making it visible to other users. 
- `edit config`: This lets you edit the client's config file. If you set 'stay logged in' to true, the client will save your refresh token and you will remain logged in across sessions. The config file should be in your .config or appdata folder, depending on whether you use linux or windows. 
- `exit`: exit


### Note:
When the program closes, including via ctrl+c, it will reset your terminal color to the default. 


